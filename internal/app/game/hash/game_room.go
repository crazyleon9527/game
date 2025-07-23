package hash

import (
	"context"
	"fmt"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service"
	"rk-api/internal/app/utils"
	"rk-api/pkg/chain"
	"rk-api/pkg/logger"
	"sync"
	"time"

	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// 回合状态机定义
const (
	RoundStatusPreparing = "preparing"
	RoundStatusBetting   = "betting"
	RoundStatusLocked    = "locked"
	RoundStatusSettling  = "settling"
	RoundStatusCompleted = "completed"
)

// 游戏房间接口
type IGameRoom interface {
	Start(context.Context)
	blockListener(context.Context)
	createNewRound()
	checkRoundProgress(uint64)
	settlementWorker(context.Context)
	processSettlement(uint64)
	retrySettlement(uint64)
	processSettlementWithBlock(uint64, *chain.Block)
	handleBlockSettlement(*GameRound, *chain.Block)
	updateRoundStatus(*GameRound)
	getLastSettledRound() *GameRound
	GetGameState() *GameState
	HandleBet(entities.IHashBetRequest) error
	processOrderAsync(entities.IHashGameOrder)

	buildHashGameRound(*entities.BaseHashGameRound) entities.IHashGameRound
	buildHashGameOrder(*entities.BaseHashGameOrder) entities.IHashGameOrder
}

// 游戏房间核心结构
type BaseGameRoom struct {
	strategy      GameStrategy // 玩法策略
	mu            sync.RWMutex
	currentRound  *GameRound
	lastRound     *GameRound
	historyRounds map[uint64]*GameRound
	blockFetcher  *chain.BlockFetcher
	settleChan    chan uint64 // 结算通道传递区块高度
	setting       *RoomSetting
	Srv           *service.HashGameService

	child IGameRoom
}

type RoomSetting struct {
	RoundInterval    uint64 // 间隔区块数（固定20）
	LockBeforeBlocks uint64 // 提前锁定区块数（5）

	Rate uint8 `gorm:"column:rate" json:"rate"` // 抽水比例
}

type GameState struct {
	LastBlockHeight uint64
	NextBlockHeight uint64
	RemainingTime   int64
	BetLockTime     time.Time
	CurrentHeight   uint64
	LastRoundResult *RoundResult
	RoundStatus     string
}

// 游戏回合结构
type GameRound struct {
	BlockHeight uint64    // 20的倍数（20,40,60...）
	StartTime   time.Time // 回合开始时间
	LockTime    time.Time // 锁定时间
	EndTime     time.Time // 结束时间
	Bets        sync.Map  // 下注记录
	Settled     bool
	Result      *RoundResult

	Status string // 新增状态字段
}

type RoundResult struct {
	BlockHeight uint64
	BlockHash   string
	Result      interface{} // 游戏结果
	Error       string      // 新增错误信息字段
}

// 初始化游戏房间
func NewBaseGameRoom(srv *service.HashGameService, strategy GameStrategy, child IGameRoom) *BaseGameRoom {
	return &BaseGameRoom{
		strategy:      strategy,
		blockFetcher:  srv.BlockFetcher,
		settleChan:    make(chan uint64, 5),
		historyRounds: make(map[uint64]*GameRound),
		setting: &RoomSetting{
			RoundInterval:    20,
			LockBeforeBlocks: 5,
		},
		Srv:   srv,
		child: child,
	}
}

func (g *BaseGameRoom) Start(ctx context.Context) {
	// 初始化首个回合
	g.createNewRound()
	go g.blockListener(ctx)
	go g.settlementWorker(ctx)
}

// 区块监听协程（增加初始回合检查）
func (g *BaseGameRoom) blockListener(ctx context.Context) {
	defer utils.PrintPanicStack()
	heightChan := g.blockFetcher.SubscribeHeight()
	for {
		select {
		case h := <-heightChan:
			// 确保当前回合存在
			g.mu.RLock()
			cr := g.currentRound
			g.mu.RUnlock()
			if cr == nil {
				g.createNewRound()
			}
			g.checkRoundProgress(h)
		case <-ctx.Done():
			return
		}
	}
}

func (g *BaseGameRoom) createNewRound() {
	g.mu.Lock()
	defer g.mu.Unlock()

	currentHeight := g.blockFetcher.GetLatestHeight()
	if currentHeight == 0 {
		logger.ZWarn("latest block height is zero, not initialize new round")
		return
	}
	targetHeight := nextRoundHeight(currentHeight)

	// 避免重复创建（例如在极短时间内多次触发）
	if g.currentRound != nil && g.currentRound.BlockHeight >= targetHeight {
		return
	}

	// 创建新回合
	r := g.child.buildHashGameRound(&entities.BaseHashGameRound{
		RoundID:     fmt.Sprintf("%d", targetHeight),
		BlockHeight: targetHeight,
		Status:      RoundStatusBetting,
	})
	if err := g.Srv.InsertHashGameRound(r); err != nil {
		logger.ZError("create new round failed", zap.Uint64("height", targetHeight), zap.Error(err))
		return
	}

	// 计算剩余区块和时间估算
	blocksRemaining := targetHeight - currentHeight
	estDuration := time.Duration(blocksRemaining*3) * time.Second

	newRound := &GameRound{
		BlockHeight: targetHeight,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(estDuration),
		LockTime:    time.Now().Add(estDuration - time.Duration(g.setting.LockBeforeBlocks*3)*time.Second),
		Status:      RoundStatusBetting,
	}

	// 转移当前回合到历史记录
	if g.currentRound != nil {
		g.historyRounds[g.currentRound.BlockHeight] = g.currentRound
		g.lastRound = g.currentRound
	}

	g.currentRound = newRound
}

// 计算下个目标高度（保持原逻辑）
func nextRoundHeight(current uint64) uint64 {
	return ((current + 20) / 20) * 20 // 更高效的整数运算
}

// 检查回合进度
func (g *BaseGameRoom) checkRoundProgress(currentHeight uint64) {
	g.mu.RLock()
	cr := g.currentRound
	g.mu.RUnlock()

	if cr == nil || cr.Settled {
		return
	}

	if cr.Status == RoundStatusBetting {
		// 判断是否到达锁定时间
		if time.Now().After(cr.LockTime) {
			cr.Status = RoundStatusLocked
		}
	}

	// 触发结算时更新状态
	if currentHeight >= cr.BlockHeight {
		cr.Status = RoundStatusSettling
		g.settleChan <- cr.BlockHeight
	}

}

// 结算处理
func (g *BaseGameRoom) settlementWorker(ctx context.Context) {
	defer utils.PrintPanicStack()
	for {
		select {
		case height := <-g.settleChan:
			g.processSettlement(height)
		case <-ctx.Done():
			return
		}
	}
}

func (g *BaseGameRoom) processSettlement(height uint64) {
	g.mu.Lock()

	round, exists := g.historyRounds[height]
	if exists && round.Settled {
		g.mu.Unlock()
		logger.ZWarn("processSettlement round already settled", zap.Uint64("height", height))
		return
	}
	round = g.currentRound

	// 标记结算开始
	round.Status = RoundStatusSettling
	g.updateRoundStatus(round)

	// 获取区块数据
	block, err := g.blockFetcher.GetBlock(height)
	if err != nil {
		g.mu.Unlock()
		g.createNewRound() // 延迟执行确保锁已释放
		go g.retrySettlement(height)
		logger.ZError("processSettlement get block failed", zap.Uint64("height", height), zap.Error(err))
		return
	}

	// 统一处理区块结算
	g.handleBlockSettlement(round, block)

	g.mu.Unlock()
	g.createNewRound() // 延迟执行确保锁已释放
}

// 重试结算逻辑
func (g *BaseGameRoom) retrySettlement(height uint64) {
	defer utils.PrintPanicStack()
	const maxRetries = 3
	retryInterval := []time.Duration{3 * time.Second, 5 * time.Second, 10 * time.Second}

	for attempt := 0; attempt < maxRetries; attempt++ {
		time.Sleep(retryInterval[attempt])

		g.mu.Lock()
		round, exists := g.historyRounds[height]
		if !exists || round.Settled {
			logger.ZWarn("retrySettlement round already settled", zap.Uint64("height", height))
			g.mu.Unlock()
			return
		}
		round.Status = RoundStatusSettling // 维持状态
		g.mu.Unlock()

		block, err := g.blockFetcher.GetBlock(height)
		if err == nil {
			logger.ZWarn("retrySettlement get block success", zap.Uint64("height", height), zap.Int("attempt", attempt+1))
			g.processSettlementWithBlock(height, block)
			return
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if round, exists := g.historyRounds[height]; exists {
		round.Result = &RoundResult{
			Error: fmt.Sprintf("结算失败，高度 %d 不可用", height),
		}
		round.Status = RoundStatusSettling // 保持状态用于人工处理
	}
}

// 带区块的结算处理（重试时调用）
func (g *BaseGameRoom) processSettlementWithBlock(height uint64, block *chain.Block) {
	g.mu.Lock()
	defer g.mu.Unlock()

	round, exists := g.historyRounds[height]
	if !exists || round.Settled {
		logger.ZWarn("processSettlementWithBlock round already settled", zap.Uint64("height", height))
		return
	}

	// 统一处理区块结算
	g.handleBlockSettlement(round, block)
}

func (g *BaseGameRoom) handleBlockSettlement(round *GameRound, block *chain.Block) {
	result := g.strategy.ParseResult(block.Hash)
	round.Result = &RoundResult{
		BlockHeight: block.Number,
		BlockHash:   block.Hash,
		Result:      result,
	}

	// 处理资金结算
	orders := make([]entities.IHashGameOrder, 0)
	round.Bets.Range(func(key, value interface{}) bool {
		order, _ := value.(entities.IHashGameOrder)
		bet := &entities.BaseHashBetRequest{
			UID:        order.GetUID(),
			BetType:    order.GetBetType(),
			BetAmount:  order.GetDelivery(),
			Prediction: order.GetPrediction(),
		}
		payout, fee := g.strategy.CalculatePayout(bet, result)
		logger.ZInfo("handleBlockSettlement betting result", zap.Any("order", order), zap.Any("result", result),
			zap.Float64("payout", payout), zap.Float64("fee", fee))
		order.SetRewardAmount(payout)
		orders = append(orders, order)
		return true
	})
	// 处理结算逻辑
	if err := g.Srv.SettlePlayerOrders(orders); err != nil {
		logger.ZError("handleBlockSettlement settle player orders failed", zap.Any("height", block.Number), zap.Error(err))
	}

	// 更新最终状态
	round.BlockHeight = block.Number
	round.Status = RoundStatusCompleted
	round.Settled = true

	// 更新数据库
	r := g.child.buildHashGameRound(&entities.BaseHashGameRound{
		RoundID: fmt.Sprintf("%d", round.BlockHeight),
		Status:  RoundStatusCompleted,
		Hash:    block.Hash,
		EndTime: time.Now().Unix(),
		Settled: 1,
		Result:  cast.ToString(result),
	})
	if err := g.Srv.UpdateHashGameRound(r); err != nil {
		logger.ZError("handleBlockSettlement update round failed", zap.Any("round", round), zap.Error(err))
	}
}

// 公共持久化方法
func (g *BaseGameRoom) updateRoundStatus(round *GameRound) {
	r := g.child.buildHashGameRound(&entities.BaseHashGameRound{
		RoundID: fmt.Sprintf("%d", round.BlockHeight),
		Status:  round.Status,
	})
	if err := g.Srv.UpdateHashGameRound(r); err != nil {
		logger.ZError("updateRoundStatus save round status failed", zap.Any("round", round), zap.Error(err))
	}
}

// 获取最后已结算的回合
func (g *BaseGameRoom) getLastSettledRound() *GameRound {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var lastRound *GameRound
	var maxHeight uint64

	for _, round := range g.historyRounds {
		if round.Settled && round.BlockHeight > maxHeight {
			maxHeight = round.BlockHeight
			lastRound = round
		}
	}
	return lastRound
}

// 前端状态接口
func (g *BaseGameRoom) GetGameState() *GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	status := &GameState{
		CurrentHeight: g.blockFetcher.GetLatestHeight(),
	}

	if g.currentRound != nil {
		status.NextBlockHeight = g.currentRound.BlockHeight
		status.RemainingTime = int64(time.Until(g.currentRound.EndTime).Seconds()) // g.currentRound.EndTime.Sub(time.Now())
		status.BetLockTime = g.currentRound.LockTime

		status.RoundStatus = g.currentRound.Status
	}

	if lastRound := g.getLastSettledRound(); lastRound != nil {
		status.LastBlockHeight = lastRound.BlockHeight
		status.LastRoundResult = lastRound.Result
	}

	return status
}

// 用户下注
func (g *BaseGameRoom) HandleBet(bet entities.IHashBetRequest) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// // 验证回合状态
	if g.currentRound == nil || g.currentRound.Status != RoundStatusBetting {
		return fmt.Errorf("HandleBet round not open")
	}

	if err := g.strategy.ValidateBet(bet); err != nil {
		return err
	}

	order := g.child.buildHashGameOrder(&entities.BaseHashGameOrder{
		RoundID:    fmt.Sprintf("%d", g.currentRound.BlockHeight),
		UID:        bet.GetUID(),
		BetTime:    time.Now().Unix(),
		BetAmount:  bet.GetBetAmount(),
		Prediction: bet.GetPrediction(),
		OrderID:    fmt.Sprintf("%d_%d", bet.GetUID(), time.Now().UnixNano()),
	})

	// 内存存储
	g.currentRound.Bets.Store(order.GetOrderID(), order)

	go g.processOrderAsync(order)

	return nil
}

func (g *BaseGameRoom) processOrderAsync(order entities.IHashGameOrder) {
	defer utils.PrintPanicStack()
	if err := g.Srv.CreateHashGameOrder(order); err != nil {
		//内存回滚
		g.currentRound.Bets.Delete(order.GetOrderID())
		logger.ZError("processOrderAsync place bet faileds", zap.Any("order", order), zap.Error(err))
		return
	}
}

func (g *BaseGameRoom) buildHashGameRound(round *entities.BaseHashGameRound) entities.IHashGameRound {
	logger.ZInfo("BaseGameRoom buildHashGameRound", zap.Any("round", round))
	return &entities.HashSDGameRound{BaseHashGameRound: round}
}

func (g *BaseGameRoom) buildHashGameOrder(order *entities.BaseHashGameOrder) entities.IHashGameOrder {
	logger.ZInfo("BaseGameRoom buildHashGameOrder", zap.Any("order", order))
	return &entities.HashSDGameOrder{BaseHashGameOrder: order}
}
