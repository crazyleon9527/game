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
	"sync"
	"time"

	"github.com/looplab/fsm"
	"go.uber.org/zap"

	"github.com/orca-zhang/ecache"
)

type OrderCache struct {
	cache *ecache.Cache
}

// 玩家订单缓存 正序排列
func NewOrderCache() *OrderCache {
	ac := &OrderCache{
		cache: ecache.NewLRUCache(20, 300, 24*time.Hour),
	}
	return ac
}

func (ac *OrderCache) add(order *entities.WingoOrder) {
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
		ac.putList(order.UID, []*entities.WingoOrder{order})
	}
}

func (ac *OrderCache) putList(uid uint, list []*entities.WingoOrder) {
	ac.cache.Put(fmt.Sprintf("%d", uid), list)
}

func (ac *OrderCache) getList(uid uint) []*entities.WingoOrder {
	if val, ok := ac.cache.Get(fmt.Sprintf("%d", uid)); ok {
		return val.([]*entities.WingoOrder)
	}
	return nil
}

func (ac *OrderCache) getLastestList(uid uint, startIndex, endIndex int) []*entities.WingoOrder {
	oldList := ac.getList(uid)
	length := len(oldList) // 旧列表长度
	// 改变索引基于正常顺序
	startIndex = length - startIndex - 1
	endIndex = length - endIndex
	// 检查索引值
	if endIndex < 0 {
		endIndex = 0
	}
	if startIndex < 0 || startIndex < endIndex {
		fmt.Println("The index value is invalid.") //无效索引值提示
		return nil
	}

	newList := make([]*entities.WingoOrder, 0, startIndex-endIndex+1) // 创建新列表

	// 从旧列表取元素，逆序添加到新列表
	for i := startIndex; i >= endIndex; i-- {
		newList = append(newList, oldList[i])
	}
	return newList
}

type PeriodHistoryList []*entities.WingoPeriod

func (s PeriodHistoryList) reverse() PeriodHistoryList {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// 新的函数，返回反转后的新数组
func (s PeriodHistoryList) reverseNew() PeriodHistoryList {
	new_s := make(PeriodHistoryList, len(s))
	copy(new_s, s)
	for i, j := 0, len(new_s)-1; i < j; i, j = i+1, j-1 {
		new_s[i], new_s[j] = new_s[j], new_s[i]
	}
	return new_s
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type WingoRoom struct {
	ID      uint
	Srv     *service.WingoService
	Setting *entities.WingoRoomSetting
	Fsm     *fsm.FSM
	// BetType             uint8
	StateSTime          int64
	RoundSTime          int64
	NowSTime            int64
	Period              *entities.WingoPeriod
	playerParticipation map[uint]uint8         //当前期玩家参与
	orders              []*entities.WingoOrder //当前期订单数
	// OrderMutex          sync.Mutex

	BettingMap     map[uint]float64
	orderCache     *OrderCache
	periodHistorys PeriodHistoryList //期数历史记录
	state          string
}

func NewWingoRoom(setting *entities.WingoRoomSetting, srv *service.WingoService) *WingoRoom {
	return &WingoRoom{
		Setting: setting,
		Srv:     srv,
		ID:      setting.ID,
	}
}

func (r *WingoRoom) init() error {
	r.periodHistorys = make([]*entities.WingoPeriod, 0)
	r.orderCache = NewOrderCache()
	list, err := r.Srv.GetLastestPeriodHistoryListWithLimit(r.Setting.BetType, PeriodHistoryMax)
	if err != nil {
		return err
	}
	r.periodHistorys = PeriodHistoryList(list)
	r.periodHistorys.reverse() //数据查出的是按插入时间倒叙,所以反转下

	return nil
}

func (r *WingoRoom) enterState(e *fsm.Event) {

	logger.ZInfo("wingo state",
		zap.String("src", e.Src),
		zap.String("Dst", e.Dst),
	)

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
func (r *WingoRoom) periodNext() error {
	// r.OrderMutex.Lock() //
	// defer r.OrderMutex.Unlock()

	r.playerParticipation = make(map[uint]uint8)
	r.BettingMap = make(map[uint]float64)
	r.orders = make([]*entities.WingoOrder, 0)
	period := &entities.WingoPeriod{
		PresetNumber: -1, //默认没设置
		// PresetNumber: 5,  //默认没设置
		Number:     -1, //默认没设置
		PeriodDate: time.Now().Format(constant.PeriodLayout),
		BetType:    r.Setting.BetType,
		Rate:       r.Setting.Rate,
		StartTime:  time.Now().Unix(),
	}
	period.EndTime = period.StartTime + int64(r.Setting.BettingInterval) + int64(r.Setting.StopBettingInterval)
	nextPeriod, err := r.Srv.CreateWingoPeriod(period)

	if err != nil {
		logger.ZPanic("create wingo period fail",
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

func (r *WingoRoom) settle(period *entities.WingoPeriod, orders []*entities.WingoOrder) {
	defer utils.PrintPanicStack()

	// r.OrderMutex.Lock() //
	// defer r.OrderMutex.Unlock()

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
	periodForUpdate := new(entities.WingoPeriod)
	structure.Copy(period, periodForUpdate)

	r.periodHistorys = append(r.periodHistorys, periodForUpdate)
	if len(r.periodHistorys) > PeriodHistoryMax+5 {
		r.periodHistorys = r.periodHistorys[5:]
	}

	logger.ZInfo("FinalizeWingoPeriod",
		zap.Any("period", periodForUpdate),
	)

	if err := r.Srv.FinalizeWingoPeriod(periodForUpdate); err != nil { //完结wingo period
		logger.ZError("UpdateWingoPeriod fail",
			zap.Any("period", period),
			zap.Error(err),
		)
	} else {
		r.Srv.SettlePlayerOrders(orders) //结算用户订单
	}

}

func (r *WingoRoom) reward(rewardNum int8, ticketNumber uint8, betAmount float64) float64 {
	return r.Srv.CalculateReward(rewardNum, ticketNumber, betAmount)
}

func (r *WingoRoom) placeBet(number uint, amount float64) {
	if currentAmount, ok := r.BettingMap[number]; ok {
		r.BettingMap[number] = currentAmount + amount
	} else {
		r.BettingMap[number] = amount
	}
}

////////////////////////////////////////////////////////////////////////////////////

func (r *WingoRoom) GetPeriodPlayerOrderList() []*entities.WingoOrder {
	list := make([]*entities.WingoOrder, 0)
	list = append(list, r.orders...)
	return list
}

// 获取最新的投注信息
func (r *WingoRoom) GetPeriodBetInfo() map[uint]float64 {
	return r.BettingMap
}

// 设置中奖号码
func (r *WingoRoom) UpdatePeriodNumber(req *entities.UpdatePeriodReq) error {
	if r.Period.PeriodID != req.PeriodID {
		return errors.With("periodID not match")
	}
	if r.state == STATE_SETTLE {
		return errors.With("current state can not update")
	}

	r.Period.Number = int8(req.Number)
	periodForUpdate := &entities.WingoPeriod{}
	periodForUpdate.ID = r.Period.ID
	periodForUpdate.Number = r.Period.Number
	return r.Srv.UpdateWingoPeriod(periodForUpdate)
}

func (r *WingoRoom) GetInfo() *entities.WingoRoomResp {
	roomInfo := &entities.WingoRoomResp{
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

func (r *WingoRoom) GetStateInfo() *entities.StateResp {
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

func (r *WingoRoom) GetRecentOrderHistoryList(req *entities.WingoOrderHistoryReq) (err error) {
	startIndex := (req.Page - 1) * req.PageSize
	endIndex := startIndex + req.PageSize // 结束index
	list := r.orderCache.getList(req.UID)
	if list == nil { //无缓存的情况
		err = r.Srv.GetLastestWingoOrderListByUID(req)
		if err != nil {
			return
		}
		r.orderCache.putList(req.UID, req.List.([]*entities.WingoOrder)) //放入缓存
		return
	} else {
		if endIndex == 1 { //返回最后一个 缓存里是按时间正序的
			req.List = []*entities.WingoOrder{list[len(list)-1]}
		} else {
			req.List = r.orderCache.getLastestList(req.UID, startIndex, endIndex)
		}
		return
	}
}

func (r *WingoRoom) GetRecentPeriodHistoryList(req *entities.GetPeriodHistoryListReq) error {
	startIndex := (req.Page - 1) * req.PageSize
	endIndex := startIndex + req.PageSize // 结束index
	list := r.periodHistorys
	// logger.Error("GetRecentPeriodHistoryList", startIndex, endIndex)
	if len(list) < endIndex { //  缓存保存的最新的数据
		return r.Srv.GetLastestPeriodHistoryList(req)
	} else {
		if endIndex == 1 { //返回最后一个 缓存里是按时间正序的
			req.List = []*entities.WingoPeriod{r.periodHistorys[len(r.periodHistorys)-1]}
		} else {
			list = r.periodHistorys.reverseNew() //逆序后全部返回
			req.List = list[startIndex:endIndex]
		}
	}
	return nil
}

func (r *WingoRoom) CreateOrder(order *entities.WingoOrderReq) (*entities.WingoOrder, error) {
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
	if r.playerParticipation[order.UID] >= PlayerOrderHistoryMax { //每期下注次数限制
		return nil, errors.WithCode(errors.UserBettingPeriodTimesLimit)
	}

	wingoOrder := new(entities.WingoOrder)
	structure.Copy(order, wingoOrder)
	wingoOrder.Rate = r.Setting.Rate
	wingoOrder.BetTime = time.Now().Unix()
	wingoOrder.FinishTime = r.Period.EndTime
	err = r.Srv.CreateWingoOrder(wingoOrder)
	if err != nil {
		return nil, err
	}

	r.placeBet(uint(wingoOrder.TicketNumber), wingoOrder.BetAmount) //添加
	r.orderCache.add(wingoOrder)                                    //加入到缓存

	r.playerParticipation[wingoOrder.UID] += 1 //map特性  不存在key时为0

	r.orders = append(r.orders, wingoOrder)

	return wingoOrder, err
}

// //////////////////////////////////////////////////////////////////////////////////////////////
// var WingoSet = wire.NewSet(
//
//	wire.Struct(new(Wingo), "Srv"),
//
// )
type IWingo interface {
	Init() error
	Start()
	GetInfo(betType uint8) *entities.WingoRoomResp
	StateSync(req *entities.StateSyncReq) *entities.StateResp
	CreateOrder(req *entities.WingoOrderReq) (*entities.WingoOrder, error)
	GetRecentOrderHistoryList(req *entities.WingoOrderHistoryReq) error
	GetRecentPeriodHistoryList(req *entities.GetPeriodHistoryListReq) error

	GetPeriodBetInfo(betType *entities.GetPeriodBetInfoReq) map[uint]float64
	GetPeriodPlayerOrderList(orderReq *entities.OrderReq) (data map[string]interface{})
	UpdatePeriodNumber(req *entities.UpdatePeriodReq) error
	GetPeriodInfo(req *entities.GetPeriodReq) *entities.PeriodInfo
	UpdateRoomLimit(req *entities.UpdateRoomLimitReq) error
}

type Wingo struct {
	Srv     *service.WingoService
	RoomMap sync.Map
}

func NewWingo(srv *service.WingoService) *Wingo {
	return &Wingo{
		Srv: srv,
	}
}

// once    sync.Once
// r.once.Do(r.init) //调用一次

func (r *Wingo) Init() error {

	r.Srv.SimulateSettleWingo() // 先把旧的未处理的期数处理完毕

	list, err := r.Srv.GetWingoSettingList()
	if err != nil {
		return errors.With(" wingo init fail" + err.Error())
	}
	for _, setting := range list {
		room := NewWingoRoom(setting, r.Srv)
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
			logger.ZError("wingo room init fail",
				zap.Uint8("bet_type", room.Setting.BetType),
				zap.Error(err),
			)
		} else {
			logger.ZInfo("wingo room init succ ",
				zap.Uint("ID", room.Setting.ID),
				zap.Uint8("bet_type", room.Setting.BetType),
			)
			r.RoomMap.Store(room.Setting.BetType, room)
		}

	}
	return nil
}

// 为0启动所有
func (r *Wingo) Start() {
	r.RoomMap.Range(func(key, value interface{}) bool {
		room, _ := value.(*WingoRoom)
		room.Fsm.Event(context.Background(), EVENT_START) //开启
		return true
	})
}

func (r *Wingo) GetRoom(betType uint8) (*WingoRoom, error) {
	// logger.Info("GetRoom:", betType)
	if room, ok := r.RoomMap.Load(betType); ok {
		return room.(*WingoRoom), nil
	}
	return nil, errors.WithCode(errors.RoomNotExist)
}

func (r *Wingo) GetInfo(betType uint8) *entities.WingoRoomResp {
	room, err := r.GetRoom(betType)
	if err != nil {
		return nil
	}
	return room.GetInfo()
}

func (r *Wingo) StateSync(req *entities.StateSyncReq) *entities.StateResp {
	// logger.Error("StateSync")
	room, err := r.GetRoom(uint8(req.BetType))
	if err != nil {
		return nil
	}
	// logger.Error("StateSync", req.PeriodID, "|", room.Period.PeriodID, "|", req.State, "|", room.state)
	if req.PeriodID == room.Period.PeriodID && room.state == req.State { //期数一样并且状态一样
		return nil
	}
	stateResp := room.GetStateInfo()
	return stateResp
}

func (r *Wingo) CreateOrder(req *entities.WingoOrderReq) (*entities.WingoOrder, error) {
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

func (r *Wingo) GetRecentPeriodHistoryList(req *entities.GetPeriodHistoryListReq) error {
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

func (r *Wingo) GetRecentOrderHistoryList(req *entities.WingoOrderHistoryReq) error {
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

func (r *Wingo) GetPeriodPlayerOrderList(orderReq *entities.OrderReq) (data map[string]interface{}) {

	list := make([]*entities.WingoOrder, 0)

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
			room := v.(*WingoRoom)
			list = append(list, room.orders...)
			return true
		})
	}

	// logger.Info("--------------GetPeriodPlayerOrderList-------------", list)

	if orderReq.PromoterCode != nil {
		filteredList := make([]*entities.WingoOrder, 0)
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
func (r *Wingo) GetPeriodBetInfo(req *entities.GetPeriodBetInfoReq) map[uint]float64 {

	room, err := r.GetRoom(req.BetType)
	if err != nil {
		return nil
	}

	return room.BettingMap

	// copiedMap := make(map[uint]float64)
	// // 遍历原map并复制键值对到新map
	// for key, value := range room.BettingMap {
	// 	copiedMap[key] = value
	// }
	// r.Srv.CalProfitInEveryNumberWithBetMap(room.orders, copiedMap) //计算每个数字下中奖盈利状况

	// logger.ZError("CalProfitInEveryNumberWithBetMap", zap.Any("cc", room.BettingMap), zap.Any("bb", copiedMap), zap.Int("len", len(copiedMap)))
	// return copiedMap
	// return map[uint]float64{
	// 	1: 300,
	// 	2: 200,
	// 	3: 10,
	// }

}

func (r *Wingo) UpdatePeriodNumber(req *entities.UpdatePeriodReq) error {
	logger.ZInfo("Wingo UpdatePeriodNumber", zap.Any("req", req))
	room, err := r.GetRoom(uint8(req.BetType))
	if err != nil {
		logger.ZError("Wingo UpdatePeriodNumber", zap.Any("req", req), zap.Error(err))
		return nil
	}

	err = room.UpdatePeriodNumber(req)
	if err != nil {
		logger.ZError("Wingo UpdatePeriodNumber", zap.Any("req", req), zap.Error(err))
		return err
	}
	return nil
}

func (r *Wingo) GetPeriodInfo(req *entities.GetPeriodReq) *entities.PeriodInfo {

	room, err := r.GetRoom(uint8(req.BetType))
	if err != nil {
		return nil
	}

	info := entities.PeriodInfo{
		PeriodID:     room.Period.PeriodID,
		PresetNumber: int8(room.Period.PresetNumber),
		DefineNumber: -1,
	}
	if room.Period.Number != -1 {
		info.DefineNumber = int8(room.Period.Number)
	}

	// logger.ZError("getp", zap.Any("info", info))

	if room.state == STATE_SETTLE {
		info.Status = 1
	} else if room.state == STATE_BETTING || room.state == STATE_INIT || room.state == STATE_WAITING {
		info.Status = 0
	}
	info.Time = time.Now().UnixMilli()
	info.CountDown = room.Period.EndTime * 1000

	info.InPcRoomLimitState = r.Srv.IsGameBetAreaLimit()
	return &info
}

func (r *Wingo) UpdateRoomLimit(req *entities.UpdateRoomLimitReq) error {

	return nil
}

// if(period.status == 1 ){ //正在结算中
// 	console.log("settle ing")
// 	triggerNext(1000)
// }else if(period.status == 0){  //正在进行中
// 	if(period.countDown < 0){//下一期还未开始
// 		triggerNext(1000)
// 	}
// 	let utilTime = Date.now() + period.countDown - period.time
// 	if(utilTime >0 ){
// 		Controller.api.djs(utilTime);
// 	}else{
// 		triggerNext(1000)
// 	}
// }
