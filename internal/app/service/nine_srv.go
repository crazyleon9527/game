package service

import (
	"fmt"
	"math/rand"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// var NineServiceSet = wire.NewSet(wire.Struct(new(NineService), "Repo", "FlowSrv", "UserSrv"))
var NineServiceSet = wire.NewSet(
	ProvideNineService,
)

type NineService struct {
	Repo                  *repository.NineRepository
	UserSrv               *UserService
	WalletSrv             *WalletService
	presetNumberListCache *ecache.Cache

	queuedOrders   map[uint][]*entities.NineOrder
	readyToProcess chan struct{}
	ordersMu       sync.Mutex
}

func ProvideNineService(
	repo *repository.NineRepository,
	userSrv *UserService,
	fundSrv *WalletService,
) *NineService {
	service := &NineService{
		Repo:           repo,
		UserSrv:        userSrv,
		WalletSrv:      fundSrv,
		queuedOrders:   make(map[uint][]*entities.NineOrder),
		readyToProcess: make(chan struct{}, 1), // 非阻塞通道
	}
	// 初始化时启动订单处理程序
	go service.processOrders()
	return service
}

func (s *NineService) processOrders() {
	for {
		// 等待新订单到达或定时触发处理
		<-s.readyToProcess
		time.Sleep(500 * time.Millisecond) // 等待1秒钟

		var toSettle []*entities.NineOrder
		s.ordersMu.Lock()
		// 遍历所有用户的订单列表
		for userID, orders := range s.queuedOrders {
			toSettle = append(toSettle, orders...)
			// 如果累计订单数达到设定的值，先进行结算
			if len(toSettle) >= 100 {
				go s.BatchSettlePlayerOrder(toSettle)
				toSettle = nil // 重置累积订单列表
			}
			delete(s.queuedOrders, userID) // 移除已处理的订单列表
		}
		s.ordersMu.Unlock()
		// 如果订单不足10个，直接结算剩余的订单
		if len(toSettle) > 0 {
			go s.BatchSettlePlayerOrder(toSettle)
		}
	}
}

var (
	NineRewardRef = map[int]float32{ //奖励设置
		1: 0.10,
		2: 0.15,
		3: 0.35,
		4: 0.6,
		5: 0.95,
		6: 1.50,
		7: 3.00,
		8: 4.00,
		9: 9.50,
	}
)

func (s *NineService) CalculateReward(rewardNum int8, ticketNumber string, betAmount float64) float64 {
	strRewardNum := fmt.Sprintf("%d", rewardNum)
	if strings.Contains(ticketNumber, strRewardNum) {
		return 0
	}
	selectArray := strings.Split(ticketNumber, ",")
	rate := NineRewardRef[len(selectArray)]
	return betAmount + betAmount*float64(rate)
}

func (s *NineService) CreateNineOrder(order *entities.NineOrder) error {
	// 在此处获取用户的锁，如果不存在，创建一个新的锁。

	user, err := s.UserSrv.GetUserByUID(order.UID)
	if err != nil {
		return err
	}

	// if user.BetAmountLimit > 0 && user.BetAmountLimit != math.MaxFloat64 {
	// 	if user.BetAmountLimit < order.BetAmount { //用户下注限制
	// 		return errors.WithCode(errors.UserBettingAmountLimit)
	// 	}
	// }

	// if user.BetTimesLimit > 0 && user.BetTimesLimit != math.MaxInt { //有每天下注限制
	// 	if s.UserSrv.CheckAndAddTodayBetTimesLmit(fmt.Sprintf("%d", user.ID), user.BetTimesLimit) {
	// 		return errors.WithCode(errors.UserDayBettingTimesLimit)
	// 	}
	// }

	wallet, err := s.WalletSrv.GetWallet(order.UID)
	if err != nil {
		return err
	}

	if wallet.Cash < order.BetAmount {
		return errors.WithCode(errors.InsufficientBalance)
	}

	err = s.WalletSrv.HandleWallet(user.ID, func(wallet *entities.UserWallet, tx *gorm.DB) error {
		wallet.SafeAdjustCash(-order.BetAmount)
		order.CalculateFee() //计算抽水
		order.Balance = wallet.Cash
		order.Username = user.Username //做上标记
		order.PromoterCode = user.PromoterCode
		order.Color = user.Color
		if err := s.Repo.CreateNineOrderWithTx(tx, order); err != nil { //创建投注单
			return err
		}

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          order.UID,
			FlowType:     constant.FLOW_TYPE_NINE,
			Number:       -order.BetAmount,
			Balance:      wallet.Cash,
			PromoterCode: user.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
		logger.ZInfo("CreateWingoOrder", zap.Any("order", order))
		return nil
	})

	return err
}

func (s *NineService) CreateNinePeriod(period *entities.NinePeriod) (*entities.NinePeriod, error) {

	periodDate := time.Now().Format(constant.PeriodLayout)
	lastestPeriod, err := s.Repo.GetLastestNinePeriodByDate(periodDate, uint8(period.BetType))
	if err != nil {
		return nil, err
	}
	if lastestPeriod == nil {
		period.PeriodDate = periodDate
		period.PeriodIndex = 1
	} else {
		period.PeriodDate = lastestPeriod.PeriodDate
		period.PeriodIndex = lastestPeriod.PeriodIndex + 1
	}
	period.PeriodID = fmt.Sprintf("%s%03d", period.PeriodDate, period.PeriodIndex)
	if period.PresetNumber == -1 { //没有设置时
		period.PresetNumber = int8(s.GetPresetNumber(periodDate, fmt.Sprintf("%d", period.BetType), int(period.PeriodIndex)))
	}
	err = s.Repo.CreateNinePeriod(period)
	if err != nil {
		return nil, err
	}
	return period, nil
}

// periodIndex  从1 开始
func (s *NineService) GetPresetNumber(periodDate string, betType string, periodIndex int) int {
	// list, _ := s.GetPresetNumberList(periodDate, betType)
	// randSrc := rand.NewSource(time.Now().UnixNano() + int64(periodIndex))
	// periodIndex = rand.New(randSrc).Intn(len(list))
	// if periodIndex > 0 && len(list) >= periodIndex {
	// 	return list[periodIndex-1]
	// }
	// return rand.New(randSrc).Intn(10)

	list, _ := s.GetPresetNumberList(periodDate, betType)
	if periodIndex > 0 && len(list) >= periodIndex {
		return list[periodIndex-1]
	}
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(10)
}

// 获取预设值列表 ，兼容旧的
func (s *NineService) GetPresetNumberList(periodDate string, betType string) ([]int, error) {
	if s.presetNumberListCache == nil {
		s.presetNumberListCache = ecache.NewLRUCache(3, 4, 24*time.Hour) //初始化缓存
	}

	cacheKey := fmt.Sprintf("%s-%s", periodDate, betType)
	if val, ok := s.presetNumberListCache.Get(cacheKey); ok {
		return val.([]int), nil
	}
	list, err := s.Repo.GetPresetNumberListRDS(periodDate, betType)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 { //预设值列表不存在
		periodDate := time.Now().Format(constant.PeriodLayout)
		list, err = s.AddPresetNumberList(periodDate, betType)
		if err != nil {
			return nil, err
		}
	}
	if len(list) > 0 {
		s.presetNumberListCache.Put(cacheKey, list) //丢入缓存中 下次从缓存读取
	}

	return list, err
}

func (s *NineService) AddPresetNumberList(periodDate string, betType string) ([]int, error) {

	bt, _ := strconv.ParseUint(betType, 10, 0)
	setting, err := s.GetNineSetting(uint(bt))
	if err != nil {
		return nil, err
	}
	roundInterval := setting.RoundInterval

	now := time.Now()
	// 当天凌晨的时间
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	// 当天结束时间的前一秒
	endOfToday := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Unix()

	var list []int
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// // 开始生成时间戳，并插入列表
	// for timestamp := startOfToday; timestamp <= endOfToday; timestamp += int64(roundInterval) {
	// 	randomNumber := r.Intn(10) //生成0-9的随机数
	// 	list = append(list, randomNumber)
	// }

	// 初始化一个真正随机数生成器
	trueRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 使用真正随机生成的数值来增强我们种子的随机性
	extraRandomness := trueRand.Int63()

	r := rand.New(rand.NewSource(time.Now().UnixNano() + extraRandomness + constant.FLOW_TYPE_NINE))
	// 开始生成时间戳，并插入列表
	for timestamp := startOfToday; timestamp <= endOfToday; timestamp += int64(roundInterval) {
		randomNumber := r.Intn(10) //生成0-9的随机数
		list = append(list, randomNumber)
	}

	// 洗牌打乱列表
	r.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	err = s.Repo.AddPresetNumberListRDS(periodDate, betType, list)
	if err != nil {
		return nil, err
	}
	return list, err
}

func (s *NineService) GetNineSettingList() ([]*entities.NineRoomSetting, error) {
	return s.Repo.GetNineSettingList()
}

func (s *NineService) GetNineSetting(betType uint) (*entities.NineRoomSetting, error) {
	return s.Repo.GetNineSetting(betType)
}

func (s *NineService) GetLastestNineOrderListByUID(param *entities.NineOrderHistoryReq) error {
	return s.Repo.GetLastestNineOrderListByUID(param)
}

func (s *NineService) GetLastestPeriodHistoryList(req *entities.GetPeriodHistoryListReq) error {
	return s.Repo.GetLastestNinePeriodList(req)
}

func (s *NineService) GetLastestPeriodHistoryListWithLimit(betType uint8, limit int) ([]*entities.NinePeriod, error) {
	return s.Repo.GetLastestPeriodHistoryListWithLimit(betType, limit)
}

func (s *NineService) FinalizeNinePeriod(period *entities.NinePeriod) error {
	period.Status = constant.STATUS_SETTLE //标记为已经处理
	if err := s.Repo.UpdateNinePeriod(period); err != nil {
		return err
	}

	logger.ZInfo("UpdateNinePeriod",
		zap.Any("period", period),
	)
	return nil
}

func (s *NineService) accumulateOrders(orders []*entities.NineOrder) error {
	s.ordersMu.Lock()
	for _, order := range orders {
		s.queuedOrders[order.UID] = append(s.queuedOrders[order.UID], order)
	}
	s.ordersMu.Unlock()
	// 非阻塞地向通道发送信号，表示有新订单
	select {
	case s.readyToProcess <- struct{}{}:
	default:
	}
	return nil
}

func (s *NineService) SettlePlayerOrders(orders []*entities.NineOrder) error {
	return s.accumulateOrders(orders)
}

// 使用事务很可能导致死锁
func (s *NineService) BatchSettlePlayerOrder(batchOrders []*entities.NineOrder) error {
	defer func() {
		if r := recover(); r != nil {
			logger.ZError("BatchSettlePlayerOrder", zap.Any("Error", r))
		}
	}()
	tx := s.Repo.DB
	for _, order := range batchOrders {
		if err := s.SettlePlayerOrderWithTx(tx, order); err != nil {
			logger.ZError("SettlePlayerOrderWithTx", zap.Any("order", order), zap.Any("Error", err))
		}
	}

	return nil
}

func (s *NineService) SettlePlayerOrderWithTx(tx *gorm.DB, order *entities.NineOrder) error {
	// s.UserSrv.Lock(order.UID)
	// defer s.UserSrv.Unlock(order.UID)

	wallet, err := s.WalletSrv.GetWallet(order.UID)
	if wallet == nil {
		return err
	}
	if order.Status == constant.STATUS_SETTLE {
		return nil
	}
	order.Status = constant.STATUS_SETTLE //标记为已经处理
	if err := s.Repo.UpdateNineOrderWithTx(tx, order); err != nil {
		return err
	}

	if order.RewardAmount > 0 {
		wallet.SafeAdjustCash(order.RewardAmount) //增加中奖金额
		// if order.FinishTime > time.Now().Unix() { //当前小于结算时间 中奖金额加到untilCash等待用户领取
		// 	user.AddUntilCash(order.RewardAmount)
		// 	user.UntilTime = order.FinishTime
		// 	userForUpdate.UntilCash = user.UntilCash
		// 	userForUpdate.UntilTime = user.UntilTime
		// } else {
		// 	walletForUpdate.Cash = wallet.Cash //直接加上
		// }

		logger.ZInfo("SettlePlayerOrderWithTx UpdateUserWithTx",
			zap.Uint("uid", wallet.ID),
			zap.Float64("balance", wallet.Cash),
			// zap.Float64("untilCash", userForUpdate.UntilCash),
			// zap.Int64("untilTime", userForUpdate.UntilTime),
		)
		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}
		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          order.UID,
			FlowType:     constant.FLOW_TYPE_NINE_REWARD,
			Number:       order.RewardAmount,
			Balance:      wallet.Cash,
			PromoterCode: wallet.PromoterCode,
		})
		if _, err := mq.MClient.Enqueue(createFlowQueue); err != nil {
			logger.ZError("createFlowQueue", zap.Any("flow", createFlowQueue), zap.Error(err))
		}
	}

	logger.ZInfo("SettlePlayerOrder", zap.Any("order", order))

	return nil
}

// 模拟结算nine 旧的期数，比如机器重启导致的
func (s *NineService) SimulateSettleNine() {

	list, err := s.Repo.GetUnSettleNinePeriodList() //获取没有处理的期数
	if err != nil {
		logger.ZError("GetUnSettleNinePeriodList fail",
			zap.Error(err),
		)
	}

	logger.ZInfo("SimulateSettleNine", zap.Any("size", len(list)))

	for _, period := range list {
		if err := s.SimulateSettleNinePeriod(period); err != nil {
			logger.ZError("SimulateSettleNinePeriod fail",
				zap.String("period", period.PeriodID),
				zap.Uint8("bet_type", period.BetType),
				zap.Error(err),
			)
		}
	}

}

func (s *NineService) SimulateSettleNineOrders(periodID string, betType uint8) error {
	period, err := s.Repo.GetNinePeriodByPeriodID(periodID, betType)
	if err != nil {
		return err
	}
	if period == nil {
		return nil
	}
	orders, err := s.Repo.GetUnSettleNineOrderListByPeriodID(period.PeriodID, period.BetType)

	if err != nil {
		return err
	}

	for _, order := range orders {
		order.Number = period.Number
		order.RewardAmount = float64(s.CalculateReward(order.Number, order.TicketNumber, order.Delivery))
	}

	for _, order := range orders {
		order.Price = period.Price
	}

	err = s.SettlePlayerOrders(orders) //结算用户订单
	return err
}

// 模拟结算单个一期
func (s *NineService) SimulateSettleNinePeriod(period *entities.NinePeriod) error {

	if period.Status == constant.STATUS_SETTLE {
		return nil
	}

	if period.Number == -1 { //如果没有设置，则为预设值
		period.Number = int8(period.PresetNumber)
	}

	orders, err := s.Repo.GetUnSettleNineOrderListByPeriodID(period.PeriodID, period.BetType)

	if err != nil {
		return err
	}

	playerParticipation := make(map[uint]uint8) //当前期玩家参与

	for _, order := range orders {
		order.Number = period.Number
		order.RewardAmount = float64(s.CalculateReward(order.Number, order.TicketNumber, order.Delivery))
		if order.RewardAmount > 0 { //玩家中奖了
			period.RewardAmount += order.RewardAmount
			period.Profit -= order.RewardAmount //盈利减去
		}
		period.Profit += order.Delivery //
		period.Fee += order.Fee
		period.BetAmount += order.BetAmount
		period.OrderCount += 1

		playerParticipation[order.UID] += 1
	}
	period.Price = float64(s.GenOpenPrice(period.RewardAmount, period.BetAmount, int(period.Number)))
	period.PlayerCount = uint(len(playerParticipation))

	for _, order := range orders {
		order.Price = period.Price
	}

	periodForUpdate := new(entities.NinePeriod)
	structure.Copy(period, periodForUpdate)

	s.FinalizeNinePeriod(periodForUpdate) //完结nine period
	s.SettlePlayerOrders(orders)          //结算用户订单

	logger.ZInfo("SimulateSettleNinePeriod succ",
		zap.String("period", period.PeriodID),
		zap.Int8("number", period.Number),
		zap.Float64("price", period.Price),
	)

	return nil
}

func (s *NineService) QuerySettleExpiredNines() error {
	order, _ := s.Repo.GetUnSettleExpiredNineOrder()
	if order == nil {
		return nil
	}
	logger.ZInfo("QuerySettleExpiredNines", zap.Any("order", order))
	return s.SimulateSettleNineOrders(order.PeriodID, order.BetType)
}

func (r *NineService) GenOpenPrice(playerWin float64, betCount float64, number int) int {
	if number < 0 {
		number = 0
	}
	first := "208"
	var second string
	three := number

	if betCount == 0 {
		// 使用种子初始化随机数生成器，通常使用当前时间作为种子
		second = fmt.Sprintf("%02d", rand.Intn(30)+10) // 生成10到39之间的随机数，包括10和39
	} else {
		val := int((playerWin / betCount) * 10)
		second = fmt.Sprintf("%02d", val%100) // 确保是两位数
	}

	openPrice, err := strconv.Atoi(first + second + strconv.Itoa(three))
	if err != nil {
		// panic(err) // 实际使用时，你可能希望处理这个错误

		return 208117
	}
	return openPrice
}

func (s *NineService) UpdateNinePeriod(period *entities.NinePeriod) error {
	return s.Repo.UpdateNinePeriod(period)
}

func (s *NineService) GetTodayTrend(param *entities.WingoTrendReq) (*entities.TrendInfo, error) {
	// list, err := s.Repo.GetTodayPeriodTrend(uint(param.BetType))
	// if err != nil {
	// 	return nil, err
	// }
	// trendInfo := &entities.TrendInfo{
	// 	Results: make([]*entities.PeriodResult, 0, len(list)),
	// }

	// for _, periodInfo := range list {
	// 	if periodInfo.Number < 0 || periodInfo.Number > 9 {
	// 		continue // no happen
	// 	}
	// 	if periodInfo.Number == 0 || periodInfo.Number == 5 {
	// 		trendInfo.VioletCount++
	// 	} else {
	// 		if periodInfo.Number < 5 {
	// 			trendInfo.GreenCount++
	// 		} else {
	// 			trendInfo.RedCount++
	// 		}
	// 	}
	// 	trendInfo.Results = append(trendInfo.Results, &entities.PeriodResult{Number: periodInfo.Number, PeriodIndex: periodInfo.PeriodIndex})
	// }

	return nil, nil
}

// 获取今天的预期  期数 状况列表
func (s *NineService) GetTodayPeriodList(req *entities.GetPeriodListReq) error {

	// logger.ZError("GetTodayPeriodList", zap.Any("req", req))

	setting, err := s.GetNineSetting(uint(req.BetType))
	if err != nil {
		return err
	}
	if setting == nil {
		return errors.With("setting not exist")
	}
	periodDate := time.Now().Format(constant.PeriodLayout)
	presetList, _ := s.GetPresetNumberList(periodDate, fmt.Sprintf("%d", req.BetType))

	list, err := s.Repo.GetTodayPeriodList(uint(req.BetType))
	if err != nil {
		return err
	}

	if len(list) <= 0 {
		return nil
	}
	if req.Page == 0 {
		req.Page = 1 // 默认从1开始
	}

	roundInterval := setting.BettingInterval + setting.StopBettingInterval

	lastestPeriod := list[len(list)-1] //正序排列，拿最近一期。 之后按正常间隔时间算。
	lastestIndex := lastestPeriod.PeriodIndex
	lastestStartTime := lastestPeriod.StartTime

	if lastestPeriod.Number == -1 {
		lastestPeriod.Number = lastestPeriod.PresetNumber //最新一期 让显示
	}

	startIndex := (req.Page - 1) * req.PageSize
	endIndex := startIndex + req.PageSize // 结束index

	newList := make([]*entities.NinePeriod, 0) //新的列表
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	midnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, now.Location()).Unix()

	var deadlineIndex int = 0
	for ; deadlineIndex < len(presetList); deadlineIndex++ { //遍历完整 presetList
		if deadlineIndex >= int(lastestIndex-1) {
			lastestStartTime += int64(roundInterval)
			if lastestStartTime > midnight {
				break
			}
		}

		if deadlineIndex >= startIndex && deadlineIndex < endIndex { //数据范围
			if deadlineIndex < int(lastestIndex) { //说明期数已经生成。
				newList = append(newList, list[deadlineIndex])
			} else {
				period := &entities.NinePeriod{
					PeriodID:     fmt.Sprintf("%s%03d", periodDate, deadlineIndex+1),
					BetType:      req.BetType,
					PresetNumber: int8(presetList[deadlineIndex]),
					StartTime:    lastestStartTime,
				}
				period.EndTime = period.StartTime + int64(roundInterval)
				period.Number = period.PresetNumber
				newList = append(newList, period)
			}
		}

	}

	// logger.Error("-----------bbbbbbbbbbbbbbbbbbbbbbb----------------", len(presetList), lastestIndex)
	// logger.Error("-----------bbbbbbbbbbbbbbbbbbbbbbb----------------", len(newList))
	// logger.Error("-----------bbbbbbbbbbbbbbbbbbbbbbb----------------", deadlineIndex, startIndex, endIndex)
	req.Count = int64(deadlineIndex)
	req.List = newList
	return nil
}
