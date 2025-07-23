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
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// var WingoServiceSet = wire.NewSet(wire.Struct(new(WingoService), "Repo", "FlowSrv", "UserSrv"))

var WingoServiceSet = wire.NewSet(
	ProvideWingoService,
)

type WingoService struct {
	Repo      *repository.WingoRepository
	UserSrv   *UserService
	AdminSrv  *AdminService
	StateSrv  *StateService
	WalletSrv *WalletService

	presetNumberListCache *ecache.Cache

	queuedOrders   map[uint][]*entities.WingoOrder
	readyToProcess chan struct{}
	ordersMu       sync.Mutex
	pcRoomlimitMap map[int]uint8
}

func ProvideWingoService(
	repo *repository.WingoRepository,
	userSrv *UserService,
	adminSrv *AdminService,
	stateSrv *StateService,
	WalletSrv *WalletService,
) *WingoService {
	service := &WingoService{
		Repo:           repo,
		UserSrv:        userSrv,
		AdminSrv:       adminSrv,
		StateSrv:       stateSrv,
		WalletSrv:      WalletSrv,
		queuedOrders:   make(map[uint][]*entities.WingoOrder),
		readyToProcess: make(chan struct{}, 1), // 非阻塞通道
		pcRoomlimitMap: make(map[int]uint8),
		// InPcRoomLimitState: false,
	}
	stateSrv.AddListener(service.onStateChange)
	// 初始化时启动订单处理程序
	go service.processOrders()
	return service
}

func (s *WingoService) processOrders() {
	for {
		// 等待新订单到达或定时触发处理
		<-s.readyToProcess
		time.Sleep(500 * time.Millisecond) // 等待1秒钟

		var toSettle []*entities.WingoOrder
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
	WingoRewardRef = map[int8]map[uint8]float32{ //奖励设置
		0: {0: 9, 20: 5.5, 30: 1.5},
		1: {1: 9, 10: 2},
		2: {2: 9, 30: 2},
		3: {3: 9, 10: 2},
		4: {4: 9, 30: 2},
		5: {5: 9, 10: 1.5, 20: 5.5},
		6: {6: 9, 30: 2},
		7: {7: 9, 10: 2},
		8: {8: 9, 30: 2},
		9: {9: 9, 10: 2},
	}
)

func (s *WingoService) onStateChange(key string, oldValue, newValue interface{}) {
	if key == constant.StateGameBetAreaLimit {
		logger.ZInfo("onStateChange", zap.String("key", key), zap.Any("oldValue", oldValue), zap.Any("newValue", newValue))
		if newValue.(bool) {
			list, _ := s.AdminSrv.GetSysUserAreaList()
			logger.ZInfo("UpdateRoomLimit", zap.Any("list", list))
			for _, v := range list {
				if v.Area >= 1 && v.Area <= 4 {
					s.pcRoomlimitMap[int(v.ID)] = uint8(v.Area)
				}
			}
		}
	}
}

func (s *WingoService) UpdateRoomLimit(req *entities.UpdateRoomLimitReq) error {
	s.StateSrv.SetState(constant.StateGameBetAreaLimit, !s.StateSrv.GetBoolState(constant.StateGameBetAreaLimit))
	return nil
}

func (s *WingoService) IsGameBetAreaLimit() bool {
	return s.StateSrv.GetBoolState(constant.StateGameBetAreaLimit)
}

func (s *WingoService) CalculateReward(rewardNum int8, ticketNumber uint8, betAmount float64) float64 {
	if ratioMap, ok := WingoRewardRef[rewardNum]; ok {
		if ratio, ok := ratioMap[ticketNumber]; ok {
			return betAmount * float64(ratio)
		}
	}
	return 0
}

func (s *WingoService) CreateWingoOrder(order *entities.WingoOrder) error {
	// 在此处获取用户的锁，如果不存在，创建一个新的锁。
	// 锁定当前用户UID，只有锁定的goroutine可以执行以下操作

	user, err := s.UserSrv.GetUserByUID(order.UID)
	if err != nil {
		return err
	}

	if s.StateSrv.GetBoolState(constant.StateGameBetAreaLimit) {
		if user.PromoterCode != 0 {
			if betType, ok := s.pcRoomlimitMap[user.PromoterCode]; ok {
				if order.BetType != betType {
					return errors.With(fmt.Sprintf("At the current time, Please bet in the designated area (G%d)", betType))
				}
			}
		}
	}

	// if user.BetAmountLimit > 0 && user.BetAmountLimit != constant.BigNumber {
	// 	if user.BetAmountLimit < order.BetAmount { //用户下注限制
	// 		return errors.WithCode(errors.UserBettingAmountLimit)
	// 	}
	// }

	// if user.BetTimesLimit > 0 && user.BetTimesLimit != constant.BigNumber { //有每天下注限制
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
		if err := s.Repo.CreateWingoOrderWithTx(tx, order); err != nil { //创建投注单
			return err
		}

		if err := s.WalletSrv.UpdateCashWithTx(tx, wallet); err != nil {
			return err
		}

		createFlowQueue, _ := handle.NewCreateFlowQueue(&entities.Flow{
			UID:          order.UID,
			FlowType:     constant.FLOW_TYPE_WINGO,
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

func (s *WingoService) CreateWingoPeriod(period *entities.WingoPeriod) (*entities.WingoPeriod, error) {

	periodDate := time.Now().Format(constant.PeriodLayout)
	lastestPeriod, err := s.Repo.GetLastestWingoPeriodByDate(periodDate, uint8(period.BetType))
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
	err = s.Repo.CreateWingoPeriod(period)
	if err != nil {
		return nil, err
	}
	return period, nil
}

// periodIndex  从1 开始
func (s *WingoService) GetPresetNumber(periodDate string, betType string, periodIndex int) int {
	list, _ := s.GetPresetNumberList(periodDate, betType)
	if periodIndex > 0 && len(list) >= periodIndex {
		return list[periodIndex-1]
	}
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(10)
}

// 获取预设值列表 ，兼容旧的
func (s *WingoService) GetPresetNumberList(periodDate string, betType string) ([]int, error) {
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
	s.presetNumberListCache.Put(cacheKey, list) //丢入缓存中 下次从缓存读取
	return list, err
}

func (s *WingoService) AddPresetNumberList(periodDate string, betType string) ([]int, error) {

	bt, _ := strconv.ParseUint(betType, 10, 0)
	setting, err := s.GetWingoSetting(uint(bt))
	if err != nil {
		return nil, err
	}
	roundInterval := setting.BettingInterval + setting.StopBettingInterval

	now := time.Now()
	// 当天凌晨的时间
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	// 当天结束时间的前一秒
	endOfToday := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Unix()

	var list []int
	// 初始化一个真正随机数生成器
	trueRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 使用真正随机生成的数值来增强我们种子的随机性
	extraRandomness := trueRand.Int63()

	r := rand.New(rand.NewSource(time.Now().UnixNano() + extraRandomness))
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

func (s *WingoService) GetWingoSettingList() ([]*entities.WingoRoomSetting, error) {
	return s.Repo.GetWingoSettingList()
}

func (s *WingoService) GetWingoSetting(betType uint) (*entities.WingoRoomSetting, error) {
	return s.Repo.GetWingoSetting(betType)
}

func (s *WingoService) GetLastestWingoOrderListByUID(param *entities.WingoOrderHistoryReq) error {
	return s.Repo.GetLastestWingoOrderListByUID(param)
}

func (s *WingoService) GetLastestPeriodHistoryList(req *entities.GetPeriodHistoryListReq) error {
	return s.Repo.GetLastestWingoPeriodList(req)
}

func (s *WingoService) GetLastestPeriodHistoryListWithLimit(betType uint8, limit int) ([]*entities.WingoPeriod, error) {
	return s.Repo.GetLastestPeriodHistoryListWithLimit(betType, limit)
}

func (s *WingoService) FinalizeWingoPeriod(period *entities.WingoPeriod) error {
	period.Status = constant.STATUS_SETTLE //标记为已经处理
	if err := s.Repo.UpdateWingoPeriod(period); err != nil {
		return err
	}

	logger.ZInfo("UpdateWingoPeriod",
		zap.Any("period", period),
	)
	return nil
}

func (s *WingoService) accumulateOrders(orders []*entities.WingoOrder) error {
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

func (s *WingoService) SettlePlayerOrders(orders []*entities.WingoOrder) error {

	return s.accumulateOrders(orders)

}

// 使用事务很可能导致死锁
func (s *WingoService) BatchSettlePlayerOrder(orders []*entities.WingoOrder) error {
	defer func() {
		if r := recover(); r != nil {
			logger.ZError("BatchSettlePlayerOrder", zap.Any("Error", r))
		}
	}()

	tx := s.Repo.DB
	for _, order := range orders {
		if err := s.SettlePlayerOrderWithTx(tx, order); err != nil {
			logger.ZError("SettlePlayerOrderWithTx", zap.Any("order", order), zap.Any("Error", err))
		}
	}
	return nil
}

func (s *WingoService) SettlePlayerOrderWithTx(tx *gorm.DB, order *entities.WingoOrder) error {

	// s.UserSrv.Lock(order.UID)
	// defer s.UserSrv.Unlock(order.UID)

	wallet, err := s.WalletSrv.GetWallet(order.UID)
	if wallet == nil {
		return err
	}
	// if err != nil {
	// 	return err
	// }
	if order.Status == constant.STATUS_SETTLE {
		return nil
	}

	order.Status = constant.STATUS_SETTLE //标记为已经处理
	if err := s.Repo.UpdateWingoOrderWithTx(tx, order); err != nil {
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
			FlowType:     constant.FLOW_TYPE_WINGO_REWARD,
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

// 模拟结算wingo 旧的期数，比如机器重启导致的
func (s *WingoService) SimulateSettleWingo() {

	list, err := s.Repo.GetUnSettleWingoPeriodList() //获取没有处理的期数
	if err != nil {
		logger.ZError("GetUnSettleWingoPeriodList fail",
			zap.Error(err),
		)
	}

	logger.ZInfo("SimulateSettleWingo", zap.Any("size", len(list)))

	for _, period := range list {
		logger.ZInfo("SimulateSettleWingo", zap.String("period", period.PeriodID))
		if err := s.SimulateSettleWingoPeriod(period); err != nil {
			logger.ZError("SimulateSettleWingoPeriod fail",
				zap.String("period", period.PeriodID),
				zap.Uint8("bet_type", period.BetType),
				zap.Error(err),
			)
		}
	}

}

func (s *WingoService) SimulateSettleWingoOrders(periodID string, betType uint8) error {
	period, err := s.Repo.GetWingoPeriodByPeriodID(periodID, betType)
	if err != nil {
		return err
	}
	if period == nil {
		return nil
	}
	orders, err := s.Repo.GetUnSettleWingoOrderListByPeriodID(period.PeriodID, period.BetType)

	if err != nil {
		return err
	}

	logger.ZInfo("SimulateSettleWingoOrders", zap.String("periodID", periodID), zap.Uint8("betType", betType), zap.Int("orders", len(orders)))

	for _, order := range orders {
		order.Number = period.Number
		order.RewardAmount = float64(s.CalculateReward(order.Number, order.TicketNumber, order.Delivery))
	}

	for _, order := range orders {
		order.Price = period.Price
	}

	err = s.BatchSettlePlayerOrder(orders) //结算用户订单
	return err
}

// 模拟结算单个一期
func (s *WingoService) SimulateSettleWingoPeriod(period *entities.WingoPeriod) error {

	if period.Status == constant.STATUS_SETTLE {
		return nil
	}

	if period.Number == -1 { //如果没有设置，则为预设值
		period.Number = int8(period.PresetNumber)
	}

	orders, err := s.Repo.GetUnSettleWingoOrderListByPeriodID(period.PeriodID, period.BetType) //获取没有结算的订单

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

	periodForUpdate := new(entities.WingoPeriod)
	structure.Copy(period, periodForUpdate)

	s.FinalizeWingoPeriod(periodForUpdate) //完结wingo period
	s.SettlePlayerOrders(orders)           //结算用户订单

	logger.ZInfo("SimulateSettleWingoPeriod succ",
		zap.String("period", period.PeriodID),
		zap.Int8("number", period.Number),
		zap.Float64("price", period.Price),
	)

	return nil
}

func (s *WingoService) QuerySettleExpiredWingos() error {
	order, _ := s.Repo.GetUnSettleExpiredWingoOrder()
	if order == nil {
		return nil
	}
	logger.ZInfo("QuerySettleExpiredWingos", zap.Any("order", order))
	return s.SimulateSettleWingoOrders(order.PeriodID, order.BetType)
}

// 计算每个数字下得盈利状况
func (s *WingoService) CalProfitInEveryNumberWithBetMap(orders []*entities.WingoOrder, betMap map[uint]float64) {
	for i := uint(0); i <= 9; i++ {
		for _, order := range orders { //you order 存在时
			betMap[40+i] += order.Delivery - float64(s.CalculateReward(int8(i), order.TicketNumber, order.Delivery))
		}
	}
}

func (r *WingoService) GenOpenPrice(playerWin float64, betCount float64, number int) int {
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

		return 208127
	}
	return openPrice
}

func (s *WingoService) UpdateWingoPeriod(period *entities.WingoPeriod) error {
	return s.Repo.UpdateWingoPeriod(period)
}

func (s *WingoService) GetTodayTrend(param *entities.WingoTrendReq) (*entities.TrendInfo, error) {
	list, err := s.Repo.GetTodayPeriodTrend(uint(param.BetType))
	if err != nil {
		return nil, err
	}
	trendInfo := &entities.TrendInfo{
		Results: make([]*entities.PeriodResult, 0, len(list)),
	}

	for _, periodInfo := range list {
		if periodInfo.Number < 0 || periodInfo.Number > 9 {
			continue // no happen
		}
		if periodInfo.Number == 0 || periodInfo.Number == 5 {
			trendInfo.VioletCount++
		} else {
			if periodInfo.Number < 5 {
				trendInfo.GreenCount++
			} else {
				trendInfo.RedCount++
			}
		}
		trendInfo.Results = append(trendInfo.Results, &entities.PeriodResult{Number: periodInfo.Number, PeriodIndex: periodInfo.PeriodIndex})
	}

	return trendInfo, nil
}

// 获取今天的预期  期数 状况列表
func (s *WingoService) GetTodayPeriodList(req *entities.GetPeriodListReq) error {

	logger.ZError("GetTodayPeriodList", zap.Any("req", req))

	setting, err := s.GetWingoSetting(uint(req.BetType))
	if err != nil {
		return err
	}
	if setting == nil {
		return errors.With("setting not exist")
	}
	periodDate := time.Now().Format(constant.PeriodLayout)
	presetList, _ := s.GetPresetNumberList(periodDate, fmt.Sprintf("%d", req.BetType))

	list, err := s.Repo.GetAllTodayPeriodList(uint(req.BetType))
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

	newList := make([]*entities.WingoPeriod, 0) //新的列表
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
				period := &entities.WingoPeriod{
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

	// logger.Error("-----------bbbbbbbbbbbbbbbbbbbbbbb----------------", len(presetList), lastestIndex, newList[0].PeriodIndex)
	// logger.Error("-----------bbbbbbbbbbbbbbbbbbbbbbb----------------", len(newList))
	// logger.Error("-----------bbbbbbbbbbbbbbbbbbbbbbb----------------", req.Page, deadlineIndex, startIndex, endIndex)
	req.Count = int64(deadlineIndex)
	req.List = newList
	return nil
}
