package crash

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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

	"github.com/google/wire"
	"go.uber.org/zap"
)

var CrashGameSet = wire.NewSet(
	NewCrashGame, // 直接提供结构体指针
)

func NewCrashGame(srv *service.CrashGameService) *CrashGame {
	m := &CrashGame{
		Srv: srv,
		// blockFetcher: srv.BlockFetcher,
		setting: &CrashSetting{
			Rate:            10,
			WaitingTimeMS:   1 * 1000,
			CountdownTimeMS: 7 * 1000,
			TakeoffTimeMS:   1.5 * 1000,
			CrashedTimeMS:   1 * 1000,
			ResultTimeMS:    6 * 1000,
			TopOrderCount:   10,
			MinBetAmount:    10,
			MaxBetAmount:    100000,
			MaxReward:       200000,

			// a: 0.00000022, // 六次项微调，为中期腾出增长空间
			// b: 0.0000035,  // 四次项略降，平衡总和
			// c: 0.00055,    // 三次项强化60%，主导中期爆发
			// d: 0.0012,     // 二次项优化，平滑衔接
			// e: 0.007,      // 一次项微调，保持前期自然
			// f: 1.0,        // 固定初始值
			a: 0.00000000035,
			b: 0.000000001,
			c: 0.0000002,
			d: 0.0044,
			e: 0.02,
			f: 1,

			BlockHash: "0000000000000000001b34dc6a1e86083f95500b096231436e9b25cbdd0075c4",
		},
		historyRounds: make(map[uint64]*GameRound),

		countdownchan: make(chan struct{}, 2),
		takeoffchan:   make(chan struct{}, 2),
		flyingchan:    make(chan struct{}, 2),
		crashedchan:   make(chan struct{}, 2),
		resultchan:    make(chan struct{}, 2),
		newroundchan:  make(chan struct{}, 2),

		orderescapechan: make(chan struct{}, 2),
	}
	m.Start(context.Background())
	return m
}

func (g *CrashGame) Start(ctx context.Context) {
	go g.crashTicker(ctx)
	go g.crashWorker(ctx)
}

func (g *CrashGame) crashTicker(ctx context.Context) {
	defer utils.PrintPanicStack()
	g.loadDBRound(ctx)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			g.checkCurrentRound(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (g *CrashGame) crashWorker(ctx context.Context) {
	defer utils.PrintPanicStack()
	for {
		select {
		case <-g.countdownchan:
			g.processCountdown(ctx)
		case <-g.takeoffchan:
			g.processTakeoff(ctx)
		case <-g.flyingchan:
			g.processFlying(ctx)
		case <-g.crashedchan:
			g.processCrashed(ctx)
		case <-g.resultchan:
			g.processResult(ctx)
		case <-g.newroundchan:
			g.processNewround(ctx)
		case <-g.orderescapechan:
			g.processOrderEscape(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (g *CrashGame) loadDBRound(_ context.Context) {
	g.Lock()
	defer g.Unlock()

	round, err := g.Srv.GetLatestCrashGameRound()
	if err != nil {
		logger.ZError("GetCrashGameRound failed", zap.Error(err))
		return
	}
	if round == nil || round.RoundID == 0 {
		g.createNewRound(1, time.Now())
		logger.ZInfo("loadDBRound createNewRound", zap.Any("round", g.currentRound))
		return
	}

	k, _ := strconv.ParseInt(round.Hash[:8], 16, 64)
	originalHash := g.genOriginalHash(round.ServerSeed, round.BlockHash)
	g.currentRound = &GameRound{
		RoundID:       round.RoundID,
		Status:        round.Status,
		ServerSeed:    round.ServerSeed,
		BlockHash:     round.BlockHash,
		OriginalHash:  originalHash,
		Hash:          round.Hash,
		OpenHash:      fmt.Sprintf("%x", sha256.Sum256([]byte(originalHash))),
		CrashK:        k,
		CrashMulti:    round.CrashMulti,
		CrashDuration: round.CrashDuration,
		WaitingTime:   time.Unix(round.WaitingTime, 0),
		Settled:       round.Settled,
		Bets:          sync.Map{},
		toporders:     make(map[uint]struct{}, 10),
	}
	g.currentRound.CountdownTime = g.currentRound.WaitingTime.Add(time.Duration(g.setting.WaitingTimeMS) * time.Millisecond)
	g.currentRound.TakeoffTime = g.currentRound.CountdownTime.Add(time.Duration(g.setting.CountdownTimeMS) * time.Millisecond)
	g.currentRound.FlyingTime = g.currentRound.TakeoffTime.Add(time.Duration(g.setting.TakeoffTimeMS) * time.Millisecond)
	g.currentRound.CrashedTime = g.currentRound.FlyingTime.Add(time.Duration(g.currentRound.CrashDuration) * time.Second)
	g.currentRound.ResultTime = g.currentRound.CrashedTime.Add(time.Duration(g.setting.CrashedTimeMS) * time.Millisecond)
	g.currentRound.NewroundTime = g.currentRound.ResultTime.Add(time.Duration(g.setting.ResultTimeMS) * time.Millisecond)

	g.nextRound = &GameRound{
		RoundID:   round.RoundID + 1,
		Status:    RoundStatusWaiting,
		Bets:      sync.Map{},
		toporders: make(map[uint]struct{}, 10),
	}

	// order
	orders, err := g.Srv.GetCrashGameOrders([]uint64{round.RoundID, round.RoundID + 1})
	if err != nil {
		logger.ZError("GetCrashGameOrders failed", zap.Error(err), zap.Uint64("roundID", round.RoundID))
		return
	}
	for _, order := range orders {
		orderID := g.buildOrderID(order.UID, order.BetIndex)
		if order.RoundID == round.RoundID {
			g.currentRound.Bets.Store(orderID, order)
		} else if order.RoundID == round.RoundID+1 {
			g.nextRound.Bets.Store(orderID, order)
		}
	}

	// auto order
	// autos, err := g.Srv.GetCrashAutoBetList(1)
	// if err != nil {
	// 	logger.ZError("loadDBRound GetCrashAutoBetList failed", zap.Error(err))
	// 	return
	// }
	// for i := range autos {
	// 	g.autoBet(autos[i])
	// }
}

func (g *CrashGame) autoBet(auto *entities.CrashAutoBet) {
	if auto.IsInfinite == 1 || auto.AutoBetCount > 0 {
		if _, err := g.PlaceCrashGameBet(&entities.PlaceCrashGameBetReq{
			UID:              auto.UID,
			BetAmount:        auto.BetAmount,
			AutoEscapeHeight: auto.AutoEscapeHeight,
		}); err == nil {
			if auto.AutoBetCount > 0 {
				auto.AutoBetCount--
			}
		}
	}
}

func (g *CrashGame) buildCurrentRound(roundID uint64, next *GameRound) *GameRound {
	current := next
	current.RoundID = roundID
	current.Status = RoundStatusWaiting
	current.BlockHash = g.setting.BlockHash
	current.OriginalHash = g.genOriginalHash(current.ServerSeed, current.BlockHash)
	current.Hash = utils.HmacSHA256(current.ServerSeed, current.BlockHash)
	current.OpenHash = fmt.Sprintf("%x", sha256.Sum256([]byte(current.OriginalHash)))
	current.CrashK, _ = strconv.ParseInt(current.Hash[:8], 16, 64)
	// max（1，2^32/（K+1）*0.99） 1%抽水
	current.CrashMulti = math.Floor(math.Max(1, (math.MaxUint32+1)/float64(current.CrashK+1)*0.99)*100) / 100
	return current
}

func (g *CrashGame) createNewRound(roundID uint64, nextStartTime time.Time) {
	g.lastRound = g.currentRound

	var current *GameRound
	if g.nextRound.ServerSeed == "" {
		g.nextRound.ServerSeed, _ = utils.GenerateSecureHex()
		current = g.buildCurrentRound(roundID, g.nextRound)
		// for count := 1; current.CrashMulti > g.setting.MaxRewardHeight && count <= 10; count++ {
		// 	if count < 10 {
		// 		g.nextRound.ServerSeed, _ = utils.GenerateSecureHex()
		// 	} else {
		// 		// 0x
		// 		g.nextRound.ServerSeed = "0f070c12204410aa7344866cc8e01b4c08c64354ea5cc3e763ec63d0335dd1d3"
		// 	}
		// 	current = g.buildCurrentRound(roundID, g.nextRound)
		// }
	} else {
		// 预设seed
		current = g.buildCurrentRound(roundID, g.nextRound)
	}
	current.WaitingTime = nextStartTime
	current.CountdownTime = current.WaitingTime.Add(time.Duration(g.setting.WaitingTimeMS) * time.Millisecond)
	current.TakeoffTime = current.CountdownTime.Add(time.Duration(g.setting.CountdownTimeMS) * time.Millisecond)
	current.FlyingTime = current.TakeoffTime.Add(time.Duration(g.setting.TakeoffTimeMS) * time.Millisecond)
	// y=ax^6+bx^4+cx^3+dx^2+ex+f
	hexEquation := utils.NewHexEquation(g.setting.a, g.setting.b, g.setting.c, g.setting.d, g.setting.e, g.setting.f)
	crashDuration, _ := hexEquation.Solve(current.CrashMulti)
	current.CrashDuration = int64(math.Round(crashDuration))
	current.CrashedTime = current.FlyingTime.Add(time.Duration(current.CrashDuration) * time.Second)
	current.ResultTime = current.CrashedTime.Add(time.Duration(g.setting.CrashedTimeMS) * time.Millisecond)
	current.NewroundTime = current.ResultTime.Add(time.Duration(g.setting.ResultTimeMS) * time.Millisecond)
	g.currentRound = current

	if err := g.Srv.CreateCrashGameRound(&entities.CrashGameRound{
		RoundID:       current.RoundID,
		Status:        current.Status,
		ServerSeed:    current.ServerSeed,
		BlockHash:     current.BlockHash,
		Hash:          current.Hash,
		CrashMulti:    current.CrashMulti,
		CrashDuration: current.CrashDuration,
		WaitingTime:   current.WaitingTime.Unix(),
		Settled:       current.Settled,
	}); err != nil {
		logger.ZError("CreateCrashGameRound failed", zap.Error(err), zap.Uint64("roundID", current.RoundID))
	}

	// notify
	g.notifyClientRound()
	g.currentRound.Bets.Range(func(key, value interface{}) bool {
		order := value.(*entities.CrashGameOrder)
		g.notifyClientOrder(order, OrderStatusBet)
		return true
	})

	// auto order
	// autos, err := g.Srv.GetCrashAutoBetList(1)
	// if err != nil {
	// 	logger.ZError("createNewRound GetCrashAutoBetList failed", zap.Error(err))
	// 	return
	// }
	// for i := range autos {
	// 	g.autoBet(autos[i])
	// }

	g.nextRound = &GameRound{
		RoundID:   roundID + 1,
		Status:    RoundStatusWaiting,
		toporders: make(map[uint]struct{}, 10),
	}
}

func (g *CrashGame) checkCurrentRound(_ context.Context) {
	g.RLock()
	defer g.RUnlock()

	switch g.currentRound.Status {
	case RoundStatusWaiting:
		if time.Now().After(g.currentRound.CountdownTime) {
			g.countdownchan <- struct{}{}
		}

	case RoundStatusCountdown:
		if time.Now().After(g.currentRound.TakeoffTime) {
			g.takeoffchan <- struct{}{}
		}

	case RoundStatusTakeoff:
		if time.Now().After(g.currentRound.FlyingTime) {
			g.flyingchan <- struct{}{}
		}

	case RoundStatusFlying:
		if time.Now().After(g.currentRound.CrashedTime) {
			g.crashedchan <- struct{}{}
		} else {
			g.orderescapechan <- struct{}{}
		}

	case RoundStatusCrashed:
		if time.Now().After(g.currentRound.ResultTime) {
			g.resultchan <- struct{}{}
		}

	case RoundStatusResult:
		// complete order load from db create new round
		if time.Now().After(g.currentRound.NewroundTime) {
			g.newroundchan <- struct{}{}
		}

	default:
		g.newroundchan <- struct{}{}
	}
}

func (g *CrashGame) notifyClientRound() {
	round, _ := json.Marshal(g.buildCrashGameRoundRsp(g.currentRound))
	g.Srv.SendMessage("all", round)
}

func (g *CrashGame) notifyClientOrder(order *entities.CrashGameOrder, status string) {
	if g.IsTopOrder(order.ID) {
		o, _ := json.Marshal(g.buildCrashGameOrderNotify(order, status))
		g.Srv.SendMessage("all", o)
	}
}

func (g *CrashGame) IsTopOrder(orderID uint) bool {
	g.currentRound.Lock()
	defer g.currentRound.Unlock()

	if _, ok := g.currentRound.toporders[orderID]; !ok {
		if len(g.currentRound.toporders) < g.setting.TopOrderCount {
			g.currentRound.toporders[orderID] = struct{}{}
			return true
		}
		return false
	}
	return true
}

func (g *CrashGame) updateStatus(status string) {
	if status != g.currentRound.Status {
		g.currentRound.Status = status
		g.Srv.UpdateCrashGameRound(&entities.CrashGameRound{
			RoundID: g.currentRound.RoundID,
			Status:  g.currentRound.Status,
			Settled: g.currentRound.Settled,
		})
	}

	g.notifyClientRound()
}

func (g *CrashGame) processCountdown(_ context.Context) {
	g.Lock()
	defer g.Unlock()

	// countdown status
	g.updateStatus(RoundStatusCountdown)
}

func (g *CrashGame) processTakeoff(_ context.Context) {
	g.Lock()
	defer g.Unlock()

	// takeoff status
	g.updateStatus(RoundStatusTakeoff)
}

func (g *CrashGame) processFlying(_ context.Context) {
	g.Lock()
	defer g.Unlock()

	// flying status
	g.updateStatus(RoundStatusFlying)
}

func (g *CrashGame) processCrashed(_ context.Context) {
	g.Lock()
	defer g.Unlock()

	// crashed status
	g.updateStatus(RoundStatusCrashed)
}

func (g *CrashGame) processResult(_ context.Context) {
	g.Lock()
	defer g.Unlock()

	// result status
	g.currentRound.Settled = constant.STATUS_SETTLE
	g.updateStatus(RoundStatusResult)

	// order settle
	orders := make([]*entities.CrashGameOrder, 0)
	g.currentRound.Bets.Range(func(key, value interface{}) bool {
		// order
		order, _ := value.(*entities.CrashGameOrder)
		if order.Status != constant.STATUS_SETTLE && order.Status != constant.STATUS_CANCEL {
			payout := g.CalculatePayoutAutoEscape(order)
			logger.ZInfo("processSettle betting result", zap.Any("order", order), zap.Any("round", g.currentRound),
				zap.Float64("payout", payout))
			order.RewardAmount = payout
			orders = append(orders, order)
		}
		return true
	})
	// 处理结算逻辑
	if err := g.Srv.SettlePlayerOrders(orders); err != nil {
		logger.ZError("processSettle settle player orders failed", zap.Any("round", g.currentRound), zap.Error(err))
	}
}

func (g *CrashGame) CalculatePayoutAutoEscape(order *entities.CrashGameOrder) float64 {
	var payout float64
	if order.EscapeHeight >= 1 && order.EscapeHeight <= g.currentRound.CrashMulti {
		payout = order.Delivery * order.EscapeHeight
	} else if g.currentRound.CrashMulti*order.Delivery >= g.setting.MaxReward {
		order.EscapeHeight = math.Floor(g.setting.MaxReward/order.Delivery*100) / 100
		payout = order.EscapeHeight * order.Delivery
	} else if order.AutoEscapeHeight >= 1 && order.AutoEscapeHeight <= g.currentRound.CrashMulti {
		order.EscapeHeight = order.AutoEscapeHeight
		payout = order.Delivery * order.AutoEscapeHeight
	}
	if payout > g.setting.MaxReward {
		payout = g.setting.MaxReward
	}
	payout = math.Round(payout*100) / 100

	return payout
}

func (g *CrashGame) CalculatePayout(order *entities.CrashGameOrder) float64 {
	var payout float64
	if order.EscapeHeight >= 1 && order.EscapeHeight <= g.currentRound.CrashMulti {
		payout = order.Delivery * order.EscapeHeight
	}
	if payout > g.setting.MaxReward {
		payout = g.setting.MaxReward
	}
	payout = math.Round(payout*100) / 100

	return payout
}

func (g *CrashGame) processNewround(_ context.Context) {
	g.Lock()
	defer g.Unlock()

	// create new round
	g.createNewRound(g.currentRound.RoundID+1, time.Now())
}

func (g *CrashGame) processOrderEscape(_ context.Context) {
	g.Lock()
	defer g.Unlock()

	past := time.Since(g.currentRound.FlyingTime).Seconds()
	hexEquation := utils.NewHexEquation(g.setting.a, g.setting.b, g.setting.c, g.setting.d, g.setting.e, g.setting.f)
	currentMultiple := hexEquation.Result(past)
	currentMultiple = math.Round(currentMultiple*100) / 100
	if currentMultiple > g.currentRound.CrashMulti {
		currentMultiple = g.currentRound.CrashMulti
	}

	orders := make([]*entities.CrashGameOrder, 0)
	g.currentRound.Bets.Range(func(key, value interface{}) bool {
		// order
		if order, ok := value.(*entities.CrashGameOrder); ok && order.EscapeHeight == 0 {
			currentReward := currentMultiple * order.Delivery
			if currentReward >= g.setting.MaxReward {
				order.EscapeHeight = math.Floor(g.setting.MaxReward/order.Delivery*100) / 100
				order.EscapeTime = time.Now().Unix()
				order.RewardAmount = math.Round(order.EscapeHeight*order.Delivery*100) / 100
				logger.ZInfo("processOrderEscape escape order", zap.Any("order", order), zap.Any("round", g.currentRound))
				orders = append(orders, order)
			} else if order.AutoEscapeHeight > 0 && order.AutoEscapeHeight <= currentMultiple {
				order.EscapeHeight = order.AutoEscapeHeight
				order.EscapeTime = time.Now().Unix()
				order.RewardAmount = math.Round(order.EscapeHeight*order.Delivery*100) / 100
				logger.ZInfo("processOrderEscape auto escape order", zap.Any("order", order), zap.Any("round", g.currentRound))
				orders = append(orders, order)
			}
		}
		return true
	})

	// 处理结算逻辑
	if err := g.Srv.SettlePlayerOrders(orders); err != nil {
		logger.ZError("processOrderEscape settle player orders failed", zap.Any("round", g.currentRound), zap.Error(err))
	} else {
		for _, order := range orders {
			g.notifyClientOrder(order, OrderStatusEscape)
		}
	}
}

func (g *CrashGame) genOriginalHash(serverSeed, blockHash string) string {
	return fmt.Sprintf("%s%s", serverSeed, blockHash)
}
