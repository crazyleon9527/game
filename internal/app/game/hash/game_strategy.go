package hash

import (
	"errors"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service"
	"sync"
)

// 游戏策略接口（合并规则和结算）
type GameStrategy interface {
	// 规则解析
	ParseResult(blockHash string) interface{}
	ResultDisplay(result interface{}) string // 结果展示格式

	// 结算处理
	CalculatePayout(bet entities.IHashBetRequest, result interface{}) (payout float64, fee float64)
	ValidateBet(bet entities.IHashBetRequest) error

	// 生命周期钩子
	BeforeSettlement(round *GameRound) // 结算前处理（如手续费计算）
	AfterSettlement(round *GameRound)  // 结算后处理（如日志记录）
}

// // 基础策略（可被嵌入）
type BaseStrategy struct {
}

func (s *BaseStrategy) Initialize() error {
	return nil
}

func (s *BaseStrategy) ParseResult(blockHash string) interface{} {
	return nil
}

func (s *BaseStrategy) ResultDisplay(result interface{}) string {
	return ""
}

func (s *BaseStrategy) CalculatePayout(bet entities.IHashBetRequest, result interface{}) (float64, float64) {
	return 0, 0
}

func (s *BaseStrategy) ValidateBet(bet entities.IHashBetRequest) error {
	return nil
}

func (s *BaseStrategy) BeforeSettlement(round *GameRound) {
}

func (s *BaseStrategy) AfterSettlement(round *GameRound) {
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func isLetter(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func findLastDigit(s string) (int, error) {
	for i := len(s) - 1; i >= 0; i-- {
		char := s[i]
		if isDigit(char) {
			return int(char - '0'), nil
		}
	}
	return 0, errors.New("未找到数字")
}

// 从后往前找到最后5位字符
func findLastFiveChars(hash string) string {
	for i := len(hash) - 1; i >= 0; i-- {
		if len(hash[i:]) >= 5 {
			return hash[i : i+5]
		}
	}
	return "" // 如果找不到5位字符，返回空字符串
}

// 解析庄家和闲家牌面并计算点数
func calculateSum(cards string) int {
	sum := 0
	for _, c := range cards {
		sum += charToValue(byte(c))
	}
	return sum % 10
}

// 字符转数值
func charToValue(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	return 0 // a/A/b/B/c/C统一为0
}

// //----------------------------------------------------------------GameRegistry-----------------------------------------------------------------------

var (
	ErrStrategyExists = errors.New("strategy already exists")
)

type GameRegistry struct {
	strategies map[GameStrategyType]GameStrategy
	rifuncs    map[GameStrategyType]func(*service.HashGameService, GameStrategy) IGameRoom
	mu         sync.RWMutex
}

func NewGameRegistry() *GameRegistry {
	return &GameRegistry{
		strategies: map[GameStrategyType]GameStrategy{
			GameStrategyTypeSingleDouble:    &SingleDoubleStrategy{},
			GameStrategyTypeSmallBig:        &SmallBigStrategy{},
			GameStrategyTypeBullBull:        &BullBullStrategy{},
			GameStrategyTypeBankerPlayerTie: &BankerPlayerTieStrategy{},
			GameStrategyTypeLucky:           &LuckyStrategy{},
		},
		rifuncs: map[GameStrategyType]func(*service.HashGameService, GameStrategy) IGameRoom{
			// GameStrategyTypeSingleDouble: NewSDGameRoom,
		},
	}
}

// 注册游戏策略
func (r *GameRegistry) Register(gameType GameStrategyType, strategy GameStrategy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.strategies[gameType]; exists {
		return ErrStrategyExists
	}

	r.strategies[gameType] = strategy
	return nil
}

// 检查玩法是否存在
func (r *GameRegistry) Exists(gameType GameStrategyType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.strategies[gameType]
	return exists
}

// 获取游戏策略
func (r *GameRegistry) GetStrategy(gameType GameStrategyType) (GameStrategy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, exists := r.strategies[gameType]
	return v, exists
}

// 获取游戏房间初始化函数
func (r *GameRegistry) GetRifunc(gameType GameStrategyType) (func(*service.HashGameService, GameStrategy) IGameRoom, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, exists := r.rifuncs[gameType]
	return v, exists
}

// // 创建策略实例
// func (r *GameRegistry) CreateStrategy(gameType string, config json.RawMessage) (GameStrategy, error) {
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()

// 	prototype, ok := r.strategies[gameType]
// 	if !ok {
// 		return nil, ErrUnknownGameType
// 	}

// 	// 克隆原型
// 	cloned := prototype.Clone()

// 	// 初始化配置
// 	if err := cloned.Initialize(config); err != nil {
// 		return nil, err
// 	}

// 	return cloned, nil
// }

// //----------------------------------------------------------------GameStrategy-----------------------------------------------------------------------

// type GameStrategy interface {
// 	Type() GameType
// 	Clone() GameStrategy
// 	Initialize(config json.RawMessage) error
// 	HandleBet(bet *entities.HashRequest) error
// 	CalculateResult(block *Block) (*entities.RoundResult, error)
// 	GetGameState() interface{}
// 	ValidateConfig(config json.RawMessage) error
// }

// // 基础策略（可被嵌入）
// type BaseStrategy struct {
// 	blockFetcher BlockFetcher
// 	room         *entities.GameRoom
// 	config       json.RawMessage
// }

// func (s *BaseStrategy) Initialize(config json.RawMessage) error {
// 	return json.Unmarshal(config, s)
// }

// func (s *BaseStrategy) ValidateConfig(config json.RawMessage) error {
// 	// 基础校验逻辑
// 	return nil
// }

// func (s *BaseStrategy) ValidateBetRequest(req entities.HashRequest) error {
// 	return nil
// }

// type RoomContext struct {
// 	Room         *entities.GameRoom
// 	currentRound *Round
// 	betChan      chan HashRequest
// 	Strategy     GameStrategy
// 	EventBus     *EventBus
// 	StopChan     chan struct{}
// }

// func (rc *RoomContext) processBet(bet *entities.HashRequest) {
// 	// 带房间ID的投注处理
// 	// 校验金额限制、人数等
// 	// ...
// }

// // 房间运行主循环
// func (rc *RoomContext) Run() {
// 	utils.PrintPanicStack()

// 	ticker := time.NewTicker(1 * time.Second)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			rc.checkRound()

// 		case bet := <-rc.betChan:
// 			rc.processBet(bet)

// 		case <-rc.cancelChan:
// 			return
// 		}
// 	}
// }

// // 在房间服务中注册事件
// func (rc *RoomContext) registerHandlers() {
// 	// 区块到达事件
// 	// rc.EventBus.Subscribe("BLOCK_ARRIVED", func(data interface{}) {
// 	// 	block := data.(*Block)
// 	// 	result := rc.strategy.CalculateResult(block)
// 	// 	rc.broadcastResult(result)
// 	// })

// 	// // 房间专属事件
// 	// rc.EventBus.Subscribe(rc.Room.ID+":CUSTOM_EVENT", func(data interface{}) {
// 	// 	// 处理自定义事件
// 	// })
// }

// // Stop 方法发送停止信号
// func (rc *RoomContext) Stop() {
// 	// 发送停止信号
// 	close(rc.StopChan)
// }
