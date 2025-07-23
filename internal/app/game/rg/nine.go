package game

import (
	"context"
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/service"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/looplab/fsm"
	"go.uber.org/zap"

	"github.com/orca-zhang/ecache"
)

type NineOrderCache struct {
	cache *ecache.Cache
}

// 玩家订单缓存 正序排列
func NewNineOrderCache() *NineOrderCache {
	ac := &NineOrderCache{
		cache: ecache.NewLRUCache(20, 300, 24*time.Hour),
	}

	return ac
}

func (ac *NineOrderCache) add(order *entities.NineOrder) {
	list := ac.getList(order.UID)
	if list != nil {
		list = append(list, order)

		// 当 list 的长度大于20时，移除最前面的5个元素
		if len(list) > PlayerOrderHistoryMax+5 {
			list = list[5:]
		}
		// 将修改后的list重新放回到缓存中
		ac.cache.Put(fmt.Sprintf("%d", order.UID), list)
	} else { // 如果 list 为 nil，建一个新的list放入缓存
		ac.putList(order.UID, []*entities.NineOrder{order})
	}
}

func (ac *NineOrderCache) putList(uid uint, list []*entities.NineOrder) {
	ac.cache.Put(fmt.Sprintf("%d", uid), list)
}

func (ac *NineOrderCache) getList(uid uint) []*entities.NineOrder {
	if val, ok := ac.cache.Get(fmt.Sprintf("%d", uid)); ok {
		return val.([]*entities.NineOrder)
	}
	return nil
}

func (ac *NineOrderCache) getLastestList(uid uint, startIndex, endIndex int) []*entities.NineOrder {
	oldList := ac.getList(uid)
	length := len(oldList) // 旧列表长度
	// 改变索引基于正常顺序
	startIndex = length - startIndex - 1
	endIndex = length - endIndex
	// 检查索引值
	if endIndex < 0 {
		endIndex = 0
	}
	// 检查索引值
	if startIndex < 0 || startIndex < endIndex {
		fmt.Println("The index value is invalid.") //无效索引值提示
		return nil
	}

	newList := make([]*entities.NineOrder, 0, startIndex-endIndex+1) // 创建新列表

	// 从旧列表取元素，逆序添加到新列表
	for i := startIndex; i >= endIndex; i-- {
		newList = append(newList, oldList[i])
	}
	return newList
}

// func (ac *NineOrderCache) dirty(uid uint) {
// 	ac.cache.Del(fmt.Sprintf("%d", uid))
// }

type NinePeriodHistoryList []*entities.NinePeriod

func (s NinePeriodHistoryList) reverse() NinePeriodHistoryList {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// 新的函数，返回反转后的新数组
func (s NinePeriodHistoryList) reverseNew() NinePeriodHistoryList {
	new_s := make(NinePeriodHistoryList, len(s))
	copy(new_s, s)
	for i, j := 0, len(new_s)-1; i < j; i, j = i+1, j-1 {
		new_s[i], new_s[j] = new_s[j], new_s[i]
	}
	return new_s
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type NineRoom struct {
	ID      uint
	Srv     *service.NineService
	Setting *entities.NineRoomSetting
	Fsm     *fsm.FSM
	// BetType             uint8
	StateSTime          int64
	RoundSTime          int64
	NowSTime            int64
	Period              *entities.NinePeriod
	playerParticipation map[uint]uint8 //当前期玩家参与

	playerSNineLimit map[uint]uint8 //当前期玩家选择9个号码次数

	orders []*entities.NineOrder //当前期订单数
	// OrderMutex sync.Mutex

	BettingMap     map[uint]float64
	orderCache     *NineOrderCache
	periodHistorys NinePeriodHistoryList //期数历史记录
	state          string
}

func NewNineRoom(setting *entities.NineRoomSetting, srv *service.NineService) *NineRoom {
	return &NineRoom{
		Setting: setting,
		Srv:     srv,
		ID:      setting.ID,
	}
}

func (r *NineRoom) init() error {
	r.periodHistorys = make([]*entities.NinePeriod, 0)
	r.orderCache = NewNineOrderCache()
	list, err := r.Srv.GetLastestPeriodHistoryListWithLimit(r.Setting.BetType, PeriodHistoryMax)
	if err != nil {
		return err
	}
	r.periodHistorys = NinePeriodHistoryList(list)
	r.periodHistorys.reverse() //数据查出的是按插入时间倒叙,所以反转下

	return nil
}

func (r *NineRoom) enterState(e *fsm.Event) {
	fmt.Printf("nine enterState %s\n", e.Dst)

	r.StateSTime = time.Now().Unix()
	r.state = r.Fsm.Current() //设置当前状态

	if e.Dst == STATE_BETTING {
		//获取setting
		r.RoundSTime = time.Now().Unix()

		if err := r.periodNext(); err != nil {
			r.Fsm.Event(context.Background(), EVENT_STOP) //停止 切换到初始状态
			return
		} else {
			time.AfterFunc(time.Duration(r.Setting.BettingInterval)*time.Second, func() {
				r.Fsm.Event(context.Background(), EVENT_STOP_BETTING)
			})
		}

	} else if e.Dst == STATE_WAITING {
		time.AfterFunc(time.Duration(r.Setting.StopBettingInterval-r.Setting.SettleInterval)*time.Second, func() {
			r.Fsm.Event(context.Background(), EVENT_SETTLE)
		})
	} else if e.Dst == STATE_SETTLE {

		go r.settle(r.Period, r.orders)
		time.AfterFunc(time.Duration(r.Setting.SettleInterval)*time.Second, func() {
			r.Fsm.Event(context.Background(), EVENT_START)
		})
	} else if e.Dst == STATE_INIT {
		//  可能因异常 切换到初始化状态
		logger.ZError(" e.Dst == STATE_INIT")
	}

}

// 开启下一轮
func (r *NineRoom) periodNext() error {
	// r.OrderMutex.Lock() //
	// defer r.OrderMutex.Unlock()

	r.playerParticipation = make(map[uint]uint8)
	r.playerSNineLimit = make(map[uint]uint8)
	r.BettingMap = make(map[uint]float64)
	r.orders = make([]*entities.NineOrder, 0)
	period := &entities.NinePeriod{
		PresetNumber: -1, //默认没设置
		Number:       -1, //默认没设置
		PeriodDate:   time.Now().Format(constant.PeriodLayout),
		BetType:      r.Setting.BetType,
		Rate:         r.Setting.Rate,
		StartTime:    time.Now().Unix(),
	}
	period.EndTime = period.StartTime + int64(r.Setting.BettingInterval) + int64(r.Setting.StopBettingInterval)

	nextPeriod, err := r.Srv.CreateNinePeriod(period)

	if err != nil {
		logger.ZPanic("create nine period fail",
			zap.String("period_date", period.PeriodDate),
			zap.Uint8("bet_type", period.BetType),
			zap.Error(err),
		)
		return err
	}
	r.Period = nextPeriod
	logger.ZInfo("periodNext", zap.Any("period", nextPeriod))
	return nil
}

func (r *NineRoom) settle(period *entities.NinePeriod, orders []*entities.NineOrder) {
	defer utils.PrintPanicStack()

	if period.Number == -1 { //如果没有设置，则为预设值
		period.Number = int8(period.PresetNumber)
	}
	period.PlayerCount = uint(len(r.playerParticipation))
	period.EndTime = time.Now().Unix()
	for _, order := range orders {
		order.Number = period.Number
		order.RewardAmount = float64(r.reward(order.Number, order.TicketNumber, order.Delivery))
		if order.RewardAmount > 0 { //玩家中奖了
			period.RewardAmount += order.RewardAmount
			period.Profit -= order.RewardAmount //盈利减去
		}

		period.Profit += order.Delivery //
		period.Fee += order.Fee
		period.BetAmount += order.BetAmount
		period.OrderCount += 1
		//
	}
	period.Price = float64(r.Srv.GenOpenPrice(period.RewardAmount, period.BetAmount, int(period.Number)))
	for _, order := range orders {
		order.Price = period.Price
	}
	periodForUpdate := new(entities.NinePeriod)
	structure.Copy(period, periodForUpdate)

	r.periodHistorys = append(r.periodHistorys, periodForUpdate)
	if len(r.periodHistorys) > PeriodHistoryMax+5 {
		r.periodHistorys = r.periodHistorys[5:]
	}

	logger.ZInfo("FinalizeWingoPeriod",
		zap.Any("period", periodForUpdate),
	)

	if err := r.Srv.FinalizeNinePeriod(periodForUpdate); err != nil { //完结nine period
		logger.ZError("UpdateNinePeriod fail",
			zap.Any("period", period),
			zap.Error(err),
		)
	} else {
		r.Srv.SettlePlayerOrders(orders) //结算用户订单
	}

}

func (r *NineRoom) reward(rewardNum int8, ticketNumber string, betAmount float64) float64 {
	return r.Srv.CalculateReward(rewardNum, ticketNumber, betAmount)
}

func (r *NineRoom) placeBet(betNumbers string, amount float64) {
	// 拆分数字字符串
	numbers := strings.Split(betNumbers, ",")

	// 遍历所有拆分出的数字
	for _, numStr := range numbers {
		// 将字符串转换为uint类型
		number, err := strconv.ParseUint(strings.TrimSpace(numStr), 10, 64)
		if err != nil {
			// 如果转换出错，可以适当处理，例如跳过该数字或输出错误
			fmt.Println("Error converting string to number:", err)
			continue
		}
		// 将转换后的数字和金额增加到BettingMap
		r.BettingMap[uint(number)] += amount
	}
}

////////////////////////////////////////////////////////////////////////////////////

func (r *NineRoom) GetPeriodPlayerOrderList() []*entities.NineOrder {
	list := make([]*entities.NineOrder, 0)
	list = append(list, r.orders...)
	return list
}

// 获取最新的投注信息
func (r *NineRoom) GetPeriodBetInfo() map[uint]float64 {
	return r.BettingMap
}

// 设置中奖号码
func (r *NineRoom) UpdatePeriodNumber(req *entities.UpdatePeriodReq) error {
	if r.Period.PeriodID != req.PeriodID {
		return errors.With("periodID not match")
	}
	if r.state == STATE_SETTLE {
		return errors.With("current state can not update")
	}

	r.Period.Number = int8(req.Number)
	periodForUpdate := &entities.NinePeriod{}
	periodForUpdate.ID = r.Period.ID
	periodForUpdate.Number = r.Period.Number
	return r.Srv.UpdateNinePeriod(periodForUpdate)
}

func (r *NineRoom) GetInfo() *entities.NineRoomResp {
	roomInfo := &entities.NineRoomResp{
		ID:          r.ID,
		PeriodIndex: r.Period.PeriodIndex,
		Setting:     r.Setting,
		PeriodID:    r.Period.PeriodID,
		StateSTime:  r.StateSTime,
		RoundSTime:  r.RoundSTime,
		NowSTime:    time.Now().Unix(),
		PlayerCount: len(r.playerParticipation),
		State:       r.state,
	}
	return roomInfo
}

func (r *NineRoom) GetStateInfo() *entities.StateResp {
	stateResp := &entities.StateResp{
		ID:          r.ID,
		PeriodIndex: r.Period.PeriodIndex,
		PeriodID:    r.Period.PeriodID,
		StateSTime:  r.StateSTime,
		RoundSTime:  r.RoundSTime,
		NowSTime:    time.Now().Unix(),
		PlayerCount: len(r.playerParticipation),
		State:       r.state,
	}
	return stateResp
}

func (r *NineRoom) GetRecentOrderHistoryList(req *entities.NineOrderHistoryReq) (err error) {
	startIndex := (req.Page - 1) * req.PageSize
	endIndex := startIndex + req.PageSize // 结束index
	list := r.orderCache.getList(req.UID)

	if list == nil { //无缓存的情况
		err = r.Srv.GetLastestNineOrderListByUID(req)
		if err != nil {
			return
		}
		r.orderCache.putList(req.UID, req.List.([]*entities.NineOrder)) //放入缓存
		return
	} else {
		if endIndex == 1 { //返回最后一个 缓存里是按时间正序的
			req.List = []*entities.NineOrder{list[len(list)-1]}
		} else {
			req.List = r.orderCache.getLastestList(req.UID, startIndex, endIndex)
		}

		return
	}
}

func (r *NineRoom) GetRecentPeriodHistoryList(req *entities.GetPeriodHistoryListReq) error {
	startIndex := (req.Page - 1) * req.PageSize
	endIndex := startIndex + req.PageSize // 结束index
	list := r.periodHistorys
	if len(list) < endIndex { //  缓存保存的最新的数据
		return r.Srv.GetLastestPeriodHistoryList(req)
	} else {
		if endIndex == 1 { //返回最后一个 缓存里是按时间正序的
			req.List = []*entities.NinePeriod{r.periodHistorys[len(r.periodHistorys)-1]}
		} else {
			// req.List = r.periodHistorys.reverseNew() //逆序后全部返回
			list = r.periodHistorys.reverseNew() //逆序后全部返回
			req.List = list[startIndex:endIndex]
		}
	}
	return nil
}

func (r *NineRoom) CreateOrder(order *entities.NineOrderReq) (*entities.NineOrder, error) {
	// r.OrderMutex.Lock() //
	// defer r.OrderMutex.Unlock()

	if r.state != STATE_BETTING {
		return nil, errors.WithCode(errors.BettingNotAllowed)
	}

	if order.PeriodID != r.Period.PeriodID {
		return nil, errors.WithCode(errors.InvalidParam)
	}
	if order.BetAmount <= 0 {
		return nil, errors.WithCode(errors.InvalidParam)
	}
	err := order.CheckTicketInvalid() //验证ticketNumber是否合法
	if err != nil {
		return nil, err
	}
	if order.IsTicketNine() { //是否选了9个数
		if r.playerSNineLimit[order.UID] > 5 { //超过5次
			return nil, errors.WithCode(errors.UserBettingPeriodTimesLimit)
		}
		r.playerSNineLimit[order.UID] += 1
	}

	if r.playerParticipation[order.UID] >= PlayerOrderHistoryMax { //每期下注次数限制
		return nil, errors.WithCode(errors.UserBettingPeriodTimesLimit)
	}

	nineOrder := new(entities.NineOrder)
	structure.Copy(order, nineOrder)
	nineOrder.Rate = r.Setting.Rate
	nineOrder.BetTime = time.Now().Unix()
	nineOrder.FinishTime = r.Period.EndTime
	err = r.Srv.CreateNineOrder(nineOrder)
	if err != nil {
		return nil, err
	}

	r.placeBet(nineOrder.TicketNumber, nineOrder.BetAmount) //添加

	r.orderCache.add(nineOrder) //加入到缓存

	r.playerParticipation[nineOrder.UID] += 1 //map特性  不存在key时为0

	r.orders = append(r.orders, nineOrder)

	return nineOrder, err
}

// //////////////////////////////////////////////////////////////////////////////////////////////
// var NineSet = wire.NewSet(
//
//	wire.Struct(new(Nine), "Srv"),
//
// )
type INine interface {
	Init() error
	Start()
	GetInfo(betType uint8) *entities.NineRoomResp
	StateSync(req *entities.StateSyncReq) *entities.StateResp
	CreateOrder(req *entities.NineOrderReq) (*entities.NineOrder, error)
	GetRecentOrderHistoryList(req *entities.NineOrderHistoryReq) error
	GetRecentPeriodHistoryList(req *entities.GetPeriodHistoryListReq) error

	GetPeriodBetInfo(req *entities.GetPeriodBetInfoReq) map[uint]float64
	GetPeriodPlayerOrderList(orderReq *entities.OrderReq) (data map[string]interface{})
	UpdatePeriodNumber(req *entities.UpdatePeriodReq) error
	GetPeriodInfo(req *entities.GetPeriodReq) *entities.PeriodInfo
}

type Nine struct {
	Srv     *service.NineService
	RoomMap sync.Map
}

func NewNine(srv *service.NineService) *Nine {
	return &Nine{
		Srv: srv,
	}
}

// once    sync.Once
// r.once.Do(r.init) //调用一次

func (r *Nine) Init() error {

	r.Srv.SimulateSettleNine() // 先把旧的未处理的期数处理完毕

	list, err := r.Srv.GetNineSettingList()
	if err != nil {
		return errors.With(" nine init fail" + err.Error())
	}
	for _, setting := range list {
		room := NewNineRoom(setting, r.Srv)
		room.Fsm = fsm.NewFSM(
			STATE_INIT,
			fsm.Events{
				{Name: EVENT_START, Src: []string{STATE_INIT, STATE_SETTLE}, Dst: STATE_BETTING},
				{Name: EVENT_STOP_BETTING, Src: []string{STATE_BETTING}, Dst: STATE_WAITING},
				{Name: EVENT_SETTLE, Src: []string{STATE_WAITING}, Dst: STATE_SETTLE},
				{Name: EVENT_STOP, Src: []string{STATE_WAITING, STATE_SETTLE, STATE_BETTING}, Dst: STATE_INIT},
			},
			fsm.Callbacks{
				"enter_state": func(_ context.Context, e *fsm.Event) { room.enterState(e) },
			},
		)
		if err := room.init(); err != nil {
			logger.ZError("nine room init fail",
				zap.Uint8("bet_type", room.Setting.BetType),
				zap.Error(err),
			)
		} else {
			logger.ZInfo("nine room init succ ",
				zap.Uint("ID", room.Setting.ID),
				zap.Uint8("bet_type", room.Setting.BetType),
			)
			r.RoomMap.Store(room.Setting.BetType, room)
		}

	}
	return nil
}

// 为0启动所有
func (r *Nine) Start() {
	defer utils.PrintPanicStack()

	r.RoomMap.Range(func(key, value interface{}) bool {
		room, _ := value.(*NineRoom)
		room.Fsm.Event(context.Background(), EVENT_START) //开启
		return true
	})
}

func (r *Nine) GetRoom(betType uint8) (*NineRoom, error) {
	if room, ok := r.RoomMap.Load(betType); ok {
		return room.(*NineRoom), nil
	}
	return nil, errors.WithCode(errors.RoomNotExist)
}

func (r *Nine) GetInfo(betType uint8) *entities.NineRoomResp {
	room, err := r.GetRoom(betType)
	if err != nil {
		return nil
	}

	info := room.GetInfo()
	return info
}

func (r *Nine) StateSync(req *entities.StateSyncReq) *entities.StateResp {
	room, err := r.GetRoom(uint8(req.BetType))
	if err != nil {
		return nil
	}
	// logger.Info("StateSync", req.PeriodID, "|", room.Period.PeriodID, "|", req.State, "|", room.state)
	if req.PeriodID == room.Period.PeriodID && room.state == req.State { //期数一样并且状态一样
		return nil
	}
	stateResp := room.GetStateInfo()
	return stateResp
}

func (r *Nine) CreateOrder(req *entities.NineOrderReq) (*entities.NineOrder, error) {
	room, err := r.GetRoom(uint8(req.BetType))
	if err != nil {
		return nil, err
	}
	order, err := room.CreateOrder(req)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *Nine) GetRecentPeriodHistoryList(req *entities.GetPeriodHistoryListReq) error {
	room, err := r.GetRoom(req.BetType)
	if err != nil {
		return err
	}
	err = room.GetRecentPeriodHistoryList(req)
	if err != nil {
		return err
	}
	return nil
}

func (r *Nine) GetRecentOrderHistoryList(req *entities.NineOrderHistoryReq) error {
	room, err := r.GetRoom(uint8(req.BetType))
	if err != nil {
		return err
	}
	err = room.GetRecentOrderHistoryList(req)
	if err != nil {
		return err
	}
	return nil
}

func (r *Nine) GetPeriodPlayerOrderList(orderReq *entities.OrderReq) (data map[string]interface{}) {

	list := make([]*entities.NineOrder, 0)

	if orderReq.BetType != 0 {
		room, err := r.GetRoom(uint8(orderReq.BetType))
		if err != nil {
			data = map[string]interface{}{
				"total": 0,
			}
			return data
		}
		list = room.GetPeriodPlayerOrderList()
	} else {
		r.RoomMap.Range(func(k, v interface{}) bool {
			room := v.(*NineRoom)
			list = append(list, room.orders...)
			return true
		})
	}

	if orderReq.PromoterCode != nil {
		filteredList := make([]*entities.NineOrder, 0)
		for _, order := range list {
			if *orderReq.PromoterCode == order.PromoterCode {
				filteredList = append(filteredList, order)
			}
		}
		list = filteredList
	}

	sort.Slice(list, func(i, j int) bool {
		switch orderReq.OrderBy {
		case 0:
			return list[i].ID > list[j].ID
		case -1:
			return list[i].BetAmount > list[j].BetAmount
		case 1:
			return list[i].BetAmount < list[j].BetAmount
		case -2:
			return list[i].Balance > list[j].Balance
		case 2:
			return list[i].Balance < list[j].Balance
		default:
			return list[i].ID > list[j].ID
		}
	})

	if orderReq.Start+orderReq.Num > len(list) {
		orderReq.Num = len(list) - orderReq.Start
	}

	result := list[orderReq.Start : orderReq.Start+orderReq.Num]
	data = map[string]interface{}{
		"list":  result,
		"total": len(list),
	}

	return data
}

// 获取最新的投注信息
func (r *Nine) GetPeriodBetInfo(req *entities.GetPeriodBetInfoReq) map[uint]float64 {

	room, err := r.GetRoom(req.BetType)
	if err != nil {
		return nil
	}
	return room.BettingMap
	// return map[uint]float64{
	// 	1: 300,
	// 	2: 200,
	// 	3: 10,
	// }

}

func (r *Nine) UpdatePeriodNumber(req *entities.UpdatePeriodReq) error {
	logger.ZInfo("Nine UpdatePeriodNumber", zap.Any("req", req))
	room, err := r.GetRoom(uint8(req.BetType))
	if err != nil {
		logger.ZError("Nine UpdatePeriodNumber", zap.Any("req", req), zap.Error(err))
		return nil
	}

	err = room.UpdatePeriodNumber(req)
	if err != nil {
		logger.ZError("Nine UpdatePeriodNumber", zap.Any("req", req), zap.Error(err))
		return err
	}
	return nil
}

func (r *Nine) GetPeriodInfo(req *entities.GetPeriodReq) *entities.PeriodInfo {
	room, err := r.GetRoom(uint8(req.BetType))
	if err != nil {
		return nil
	}

	// logger.ZError("GetPeriodInfo", zap.Any("req", req), zap.Any("room", room.Period))

	info := entities.PeriodInfo{
		PeriodID:     room.Period.PeriodID,
		PresetNumber: int8(room.Period.PresetNumber),
		DefineNumber: -1,
	}
	if room.Period.Number != -1 {
		info.DefineNumber = int8(room.Period.Number)
	}

	if room.state == STATE_SETTLE {
		info.Status = 1
	} else if room.state == STATE_BETTING || room.state == STATE_INIT || room.state == STATE_WAITING {
		info.Status = 0
	}
	info.Time = time.Now().UnixMilli()
	info.CountDown = room.Period.EndTime * 1000

	return &info
}
