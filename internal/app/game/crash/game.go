package crash

import (
	"crypto/sha256"
	"fmt"
	"math"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// 回合状态机定义
const (
	RoundStatusWaiting   = "waiting"   // 等待 1s
	RoundStatusCountdown = "countdown" // 下注 7s
	RoundStatusTakeoff   = "takeoff"   // 起飞 1.5s
	RoundStatusFlying    = "flying"    // 飞行 实时计算时间
	RoundStatusCrashed   = "crashed"   // 结算 1s
	RoundStatusResult    = "result"    // 结束 6s
)

// OrderStatus
const (
	OrderStatusBet       = "bet"
	OrderStatusCancelBet = "cancelbet"
	OrderStatusEscape    = "escape"
)

// notify type
const (
	NotifyTypeRound = "round"
	NotifyTypeOrder = "order"
)

type CrashGame struct {
	Srv *service.CrashGameService
	// blockFetcher  *chain.BlockFetcher
	setting *CrashSetting

	sync.RWMutex
	currentRound  *GameRound
	lastRound     *GameRound
	nextRound     *GameRound
	historyRounds map[uint64]*GameRound

	countdownchan chan struct{} // 下注
	takeoffchan   chan struct{} // 起飞
	flyingchan    chan struct{} // 飞行
	crashedchan   chan struct{} // 爆炸
	resultchan    chan struct{} // 结果
	newroundchan  chan struct{} // 新回合

	orderescapechan chan struct{} // 订单逃逸
}

// 游戏回合结构
type GameRound struct {
	RoundID       uint64
	Status        string // 状态字段
	ServerSeed    string
	BlockHash     string
	OriginalHash  string
	Hash          string
	OpenHash      string
	CrashK        int64
	CrashMulti    float64
	CrashDuration int64
	WaitingTime   time.Time // 等待时间
	CountdownTime time.Time // 下注时间
	TakeoffTime   time.Time // 起飞时间
	FlyingTime    time.Time // 飞行时间
	CrashedTime   time.Time // 爆炸时间
	ResultTime    time.Time // 结果时间
	NewroundTime  time.Time // 新回合时间
	Settled       uint8

	Bets sync.Map // 下注记录

	sync.RWMutex
	toporders map[uint]struct{}
}

type CrashSetting struct {
	// 抽水比例 1%
	Rate uint8
	// 等待	1S
	WaitingTimeMS int64
	// 下注	7S
	CountdownTimeMS int64
	// 起飞	1.5S
	TakeoffTimeMS int64
	// 爆炸	1S
	CrashedTimeMS int64
	// 结果 6S
	ResultTimeMS int64
	// 下注订单展示数量
	TopOrderCount int
	// 最小下注 10
	MinBetAmount float64
	// 最大下注 100000
	MaxBetAmount float64
	// 最大奖金 200000
	MaxReward float64
	// x=ax^6+bx^4+cx^3+dx^2+ex+f 参数
	a, b, c, d, e, f float64
	// 比特币区块 584500 的哈希值固定为 0000000000000000001b34dc6a1e86083f95500b096231436e9b25cbdd0075c4
	BlockHash string
}

func (g *CrashGame) GetCrashGameRound() (*entities.GetCrashGameRoundRsp, error) {
	g.RLock()
	defer g.RUnlock()

	return g.buildCrashGameRoundRsp(g.currentRound), nil
}

func (g *CrashGame) buildCrashGameRoundRsp(round *GameRound) *entities.GetCrashGameRoundRsp {
	seed, hash, crashMulti := "", "", float64(0)
	if round.Status == RoundStatusCrashed || round.Status == RoundStatusResult {
		seed, hash, crashMulti = round.ServerSeed, round.Hash, round.CrashMulti
	}
	currentStatusTime := g.getCurrentStatusTime(round)
	return &entities.GetCrashGameRoundRsp{
		RoundID:           round.RoundID,
		Status:            round.Status,
		ServerSeed:        seed,
		Hash:              hash,
		OpenHash:          round.OpenHash,
		CrashMulti:        crashMulti,
		CurrentTime:       time.Now().Unix(),
		CurrentStatusTime: currentStatusTime,
		Settled:           round.Settled,
		NotifyType:        NotifyTypeRound,
	}
}

func (g *CrashGame) getCurrentStatusTime(round *GameRound) int64 {
	switch round.Status {
	case RoundStatusWaiting:
		return round.WaitingTime.Unix()
	case RoundStatusCountdown:
		return round.CountdownTime.Unix()
	case RoundStatusTakeoff:
		return round.TakeoffTime.Unix()
	case RoundStatusFlying:
		return round.FlyingTime.Unix()
	case RoundStatusCrashed:
		return round.CrashedTime.Unix()
	case RoundStatusResult:
		return round.ResultTime.Unix()
	default:
		return 0
	}
}

func (g *CrashGame) GetDBCrashGameRound(roundID uint64) (*entities.GetCrashGameRoundRsp, error) {
	g.RLock()
	defer g.RUnlock()
	if roundID == g.currentRound.RoundID {
		return g.buildCrashGameRoundRsp(g.currentRound), nil
	}

	round, err := g.Srv.GetCrashGameRound(roundID)
	if err != nil {
		return nil, err
	}
	order, err := g.Srv.GetTopHeightCrashGameOrder(round.RoundID)
	if err != nil {
		return nil, err
	}
	if order.EscapeHeight > 0 {
		return g.buildDBCrashGameRoundRsp(round, order.Name, order.EscapeHeight), nil
	}
	return g.buildCrashGameRoundRsp(g.currentRound), nil
}

func (g *CrashGame) buildDBCrashGameRoundRsp(round *entities.CrashGameRound, name string, escapeHeight float64) *entities.GetCrashGameRoundRsp {
	originalHash := g.genOriginalHash(round.ServerSeed, round.BlockHash)
	return &entities.GetCrashGameRoundRsp{
		RoundID:     round.RoundID,
		Status:      round.Status,
		ServerSeed:  round.ServerSeed,
		Hash:        round.Hash,
		OpenHash:    fmt.Sprintf("%x", sha256.Sum256([]byte(originalHash))),
		CrashMulti:  round.CrashMulti,
		CurrentTime: time.Now().Unix(),
		Settled:     round.Settled,

		UltimateEscaper: name,
		ExtremeAltitude: escapeHeight,
	}
}

func (g *CrashGame) GetCrashGameRoundList() (*entities.GetCrashGameRoundListRsp, error) {
	list, err := g.Srv.GetCrashGameRoundList()
	if err != nil {
		return nil, err
	}
	rids := make([]uint64, 0, len(list))
	for _, round := range list {
		rids = append(rids, round.RoundID)
	}

	orders, err := g.Srv.GetTopHeightCrashGameOrderList(rids)
	if err != nil {
		return nil, err
	}
	orderMap := make(map[uint64]*entities.CrashGameOrder, len(orders))
	for _, order := range orders {
		orderMap[order.RoundID] = order
	}

	avarageCount, avarageEscapeAltitude := float64(0), float64(0)
	rsp := &entities.GetCrashGameRoundListRsp{List: make([]*entities.GetCrashGameRoundRsp, 0, len(list))}
	for _, round := range list {
		if order, ok := orderMap[round.RoundID]; ok && order.EscapeHeight > 0 {
			avarageCount, avarageEscapeAltitude = avarageCount+1, avarageEscapeAltitude+order.EscapeHeight
			rsp.List = append(rsp.List, g.buildDBCrashGameRoundRsp(round, order.Name, order.EscapeHeight))
		} else {
			rsp.List = append(rsp.List, g.buildDBCrashGameRoundRsp(round, "", 0))
		}
	}
	if avarageCount == 0 {
		rsp.AverageEscapeAltitude = 0
	} else {
		rsp.AverageEscapeAltitude = math.Round(avarageEscapeAltitude/avarageCount*100) / 100
	}
	return rsp, nil
}

func (g *CrashGame) GetCrashGameRoundOrderList() ([]*entities.CrashGameOrder, error) {
	g.RLock()
	defer g.RUnlock()

	var orders []*entities.CrashGameOrder
	g.currentRound.Bets.Range(func(key, value interface{}) bool {
		order, _ := value.(*entities.CrashGameOrder)
		order.RewardAmount = g.CalculatePayout(order)
		orders = append(orders, order)
		return true
	})
	return orders, nil
}

func (g *CrashGame) GetDBCrashGameRoundOrderList(roundID uint64) ([]*entities.CrashGameOrder, error) {
	g.RLock()
	defer g.RUnlock()
	if roundID == g.currentRound.RoundID {
		var orders []*entities.CrashGameOrder
		g.currentRound.Bets.Range(func(key, value interface{}) bool {
			order, _ := value.(*entities.CrashGameOrder)
			order.RewardAmount = g.CalculatePayout(order)
			orders = append(orders, order)
			return true
		})
		return orders, nil
	}
	return g.Srv.GetCrashGameOrders([]uint64{roundID})
}

func (g *CrashGame) buildCrashGameOrderNotify(order *entities.CrashGameOrder, status string) *entities.CrashGameOrderNotify {
	rewardAmount := order.RewardAmount
	return &entities.CrashGameOrderNotify{
		UID:          order.UID,
		Name:         order.Name,
		RoundID:      order.RoundID,
		BetIndex:     order.BetIndex,
		BetAmount:    order.BetAmount,
		BetTime:      order.BetTime,
		EscapeHeight: order.EscapeHeight,
		EscapeTime:   order.EscapeTime,
		RewardAmount: rewardAmount,
		Status:       status,
		NotifyType:   NotifyTypeOrder,
	}
}

func (g *CrashGame) PlaceCrashGameBet(req *entities.PlaceCrashGameBetReq) (*entities.CrashGameOrder, error) {
	g.Lock()
	defer g.Unlock()

	// 验证回合状态
	if err := g.validateRound(); err != nil {
		return nil, err
	}
	// 验证下注信息
	if err := g.validateBet(req); err != nil {
		return nil, err
	}

	// 内存存储
	order := &entities.CrashGameOrder{
		UID:              req.UID,
		BetIndex:         req.BetIndex,
		AutoEscapeHeight: req.AutoEscapeHeight,
		// Rate:             g.setting.Rate,
		BetTime:   time.Now().Unix(),
		BetAmount: req.BetAmount,

		OrderID: g.buildOrderID(req.UID, req.BetIndex),
	}
	if g.currentRound.Status == RoundStatusCountdown || g.currentRound.Status == RoundStatusWaiting {
		if _, ok := g.currentRound.Bets.Load(order.OrderID); ok {
			return nil, fmt.Errorf("PlaceCrashGameBet current round order exist")
		}
		order.RoundID = g.currentRound.RoundID
		g.currentRound.Bets.Store(order.OrderID, order)
	} else {
		if _, ok := g.nextRound.Bets.Load(order.OrderID); ok {
			return nil, fmt.Errorf("PlaceCrashGameBet next round order exist")
		}
		order.RoundID = g.currentRound.RoundID + 1
		g.nextRound.Bets.Store(order.OrderID, order)
	}

	// 异步处理
	go g.processOrderAsync(order)
	return order, nil
}

func (g *CrashGame) validateRound() error {
	if g.currentRound == nil || g.nextRound == nil {
		return fmt.Errorf("validateRound round not open")
	}
	return nil
}

func (g *CrashGame) validateBet(req *entities.PlaceCrashGameBetReq) error {
	if req.BetAmount < g.setting.MinBetAmount || req.BetAmount > g.setting.MaxBetAmount {
		return fmt.Errorf("validateBet params error")
	}
	return nil
}

func (g *CrashGame) buildOrderID(uid uint, betIndex int) string {
	return fmt.Sprintf("%d-%d", uid, betIndex)
}

func (g *CrashGame) processOrderAsync(order *entities.CrashGameOrder) {
	defer utils.PrintPanicStack()
	if err := g.Srv.CreateCrashGameOrder(order); err != nil {
		g.Lock()
		defer g.Unlock()
		//内存回滚
		if order.RoundID == g.currentRound.RoundID {
			g.currentRound.Bets.Delete(order.OrderID)
		} else {
			g.nextRound.Bets.Delete(order.OrderID)
		}
		logger.ZError("processOrderAsync place bet faileds", zap.Any("order", order), zap.Error(err))
		return
	} else {
		if order.RoundID == g.currentRound.RoundID {
			g.notifyClientOrder(order, OrderStatusBet)
		}
	}
}

func (g *CrashGame) CancelCrashGameBet(req *entities.CancelCrashGameBetReq) error {
	g.Lock()
	defer g.Unlock()

	// 验证回合状态
	if err := g.validateRound(); err != nil {
		return err
	}

	// 取消订单
	var err error
	orderID := g.buildOrderID(req.UID, req.BetIndex)
	if g.currentRound.Status == RoundStatusCountdown || g.currentRound.Status == RoundStatusWaiting {
		if v, ok := g.currentRound.Bets.Load(orderID); ok {
			order := v.(*entities.CrashGameOrder)
			if err = g.processCancelOrder(order); err == nil {
				g.currentRound.Bets.Delete(orderID)
				g.notifyClientOrder(order, OrderStatusCancelBet)
			}
		} else {
			return fmt.Errorf("PlaceCrashGameBet current round order not exist")
		}
	} else {
		if v, ok := g.nextRound.Bets.Load(orderID); ok {
			order := v.(*entities.CrashGameOrder)
			if err = g.processCancelOrder(order); err == nil {
				g.nextRound.Bets.Delete(orderID)
			}
		} else {
			return fmt.Errorf("PlaceCrashGameBet next round order not exist")
		}
	}

	return err
}

func (g *CrashGame) processCancelOrder(order *entities.CrashGameOrder) error {
	if order.Status == constant.STATUS_SETTLE || order.Status == constant.STATUS_CANCEL {
		return fmt.Errorf("processCancelOrder order status error")
	}
	order.Status = constant.STATUS_CANCEL
	return g.Srv.CancelCrashGameOrder(order)
}

func (g *CrashGame) EscapeCrashGameBet(req *entities.EscapeCrashGameBetReq) (*entities.CrashGameOrder, error) {
	g.Lock()
	defer g.Unlock()

	// 验证回合状态
	if err := g.validateRound(); err != nil {
		return nil, err
	}
	// 验证逃跑参数
	if err := g.validateEscape(req); err != nil {
		return nil, err
	}

	// 逃跑
	orderID := g.buildOrderID(req.UID, req.BetIndex)
	if g.currentRound.Status == RoundStatusFlying {
		if v, ok := g.currentRound.Bets.Load(orderID); ok {
			order := v.(*entities.CrashGameOrder)
			if err := g.processEscapeOrder(order, req); err != nil {
				return nil, err
			}
			g.notifyClientOrder(order, OrderStatusEscape)
			return order, nil
		} else {
			return nil, fmt.Errorf("EscapeCrashGameBet current round order not exist")
		}
	} else {
		return nil, fmt.Errorf("EscapeCrashGameBet round status error")
	}
}

func (g *CrashGame) validateEscape(req *entities.EscapeCrashGameBetReq) error {
	if req.EscapeHeight <= 1 || req.EscapeHeight > g.currentRound.CrashMulti {
		return fmt.Errorf("validateEscape params EscapeHeight not in range")
	}
	return nil
}

func (g *CrashGame) processEscapeOrder(order *entities.CrashGameOrder, req *entities.EscapeCrashGameBetReq) error {
	if order.Status == constant.STATUS_SETTLE || order.Status == constant.STATUS_CANCEL {
		return fmt.Errorf("processEscapeOrder order status error")
	}
	if order.EscapeHeight > 0 || order.EscapeTime > 0 {
		return fmt.Errorf("processEscapeOrder order already escape error")
	}
	// if order.EscapeTime > time.Now().Unix() ||
	// 	order.EscapeTime < g.currentRound.EndBetTime.Unix() ||
	// 	order.EscapeTime >= g.currentRound.CrashTime.Unix() {
	// 	return fmt.Errorf("processEscapeOrder escape time error")
	// }
	// // 校验逃跑高度是否正确
	// hexEquation := utils.NewHexEquation(g.setting.a, g.setting.b, g.setting.c, g.setting.d, g.setting.e, g.setting.f)
	// crashTime, _ := hexEquation.Solve(order.EscapeHeight)
	// escapeTime := order.EscapeTime - g.currentRound.StartBetTime.Unix()
	// if int64(crashTime) != escapeTime {
	// 	return fmt.Errorf("processEscapeOrder escape height error")
	// }
	order.EscapeHeight = req.EscapeHeight
	order.EscapeTime = req.EscapeTime
	order.RewardAmount = g.CalculatePayout(order)
	return g.Srv.SettleOrder(order)
}

func (g *CrashGame) GetCrashAutoBet(uid uint) (*entities.CrashAutoBet, error) {
	return g.Srv.GetCrashAutoBet(uid)
}

func (g *CrashGame) PlaceCrashAutoBet(req *entities.PlaceCrashAutoBetReq) (*entities.CrashAutoBet, error) {
	bet := &entities.CrashAutoBet{
		UID:              req.UID,
		BetAmount:        req.BetAmount,
		AutoEscapeHeight: req.AutoEscapeHeight,
		AutoBetCount:     req.AutoBetCount,
		Status:           1,
	}
	if req.AutoBetCount == 0 {
		bet.IsInfinite = 1
	}
	// 创建订单
	g.autoBet(bet)
	// 创建自动逃跑
	if err := g.Srv.CreateCrashAutoBet(bet); err != nil {
		return nil, err
	}
	return bet, nil
}

func (g *CrashGame) CancelCrashAutoBet(uid uint) error {
	return g.Srv.UpdateCrashAutoBetStatus(uid, 0)
}

func (g *CrashGame) GetUserCrashGameOrder(uid uint) ([]*entities.CrashGameOrder, error) {
	g.RLock()
	defer g.RUnlock()

	var orders []*entities.CrashGameOrder
	g.currentRound.Bets.Range(func(key, value interface{}) bool {
		if v, ok := value.(*entities.CrashGameOrder); ok && v.UID == uid {
			orders = append(orders, v)
		}
		return true
	})
	return orders, nil
}

func (g *CrashGame) GetDBUserCrashGameOrder(uid uint, roundID uint64) ([]*entities.CrashGameOrder, error) {
	g.RLock()
	defer g.RUnlock()
	if roundID == g.currentRound.RoundID {
		var orders []*entities.CrashGameOrder
		g.currentRound.Bets.Range(func(key, value interface{}) bool {
			if v, ok := value.(*entities.CrashGameOrder); ok && v.UID == uid {
				orders = append(orders, v)
			}
			return true
		})
		return orders, nil
	}

	return g.Srv.GetUserCrashGameOrder(uid, roundID)
}

func (g *CrashGame) GetUserCrashGameOrderList(uid uint) (*entities.GetUserCrashGameOrderListRsp, error) {
	list, err := g.Srv.GetUserCrashGameOrderList(uid)
	if err != nil {
		return nil, err
	}
	totalBetAmount, totalRewardAmount := float64(0), float64(0)
	for i := range list {
		totalBetAmount += list[i].BetAmount
		totalRewardAmount += list[i].RewardAmount
	}

	return &entities.GetUserCrashGameOrderListRsp{
		List:              list,
		TotalBetAmount:    totalBetAmount,
		TotalRewardAmount: totalRewardAmount,
	}, nil
}

func FairCheck(req *entities.FairCheckReq) (*entities.FairCheckRsp, error) {
	hash := utils.HmacSHA256(req.ServerSeed, req.BlockHash)
	crashK, _ := strconv.ParseInt(hash[:8], 16, 64)
	// max（1，2^32/（K+1）*0.99）
	crashMulti := math.Floor(math.Max(1, (math.MaxUint32+1)/float64(crashK+1)*0.99)*100) / 100
	return &entities.FairCheckRsp{Result: crashMulti}, nil
}

func (g *CrashGame) Test(seed string) {
	g.Lock()
	defer g.Unlock()

	g.nextRound.ServerSeed = seed
}
