package entities

import (
	"rk-api/pkg/math"

	"github.com/shopspring/decimal"
)

// -------------------------------- sql --------------------------------

type CrashGameRound struct {
	BaseModel
	RoundID       uint64  `gorm:"column:round_id;size:35;uniqueIndex:idx_round_id" json:"round_id"` // 轮数：每进行一局，轮数+1
	Status        string  `gorm:"column:status" json:"status"`                                      // 状态: preparing:准备,上一局结算时间5.5s betting:下注7s flying:飞行,具体时间根据爆炸倍数计算 settling:结算 completed:已完成
	ServerSeed    string  `gorm:"column:server_seed" json:"server_seed"`                            // 服务器种子：由系统随机生成，一般为64位
	BlockHash     string  `gorm:"column:block_hash" json:"block_hash"`                              // 比特币区块哈希：取最新的区块哈希
	Hash          string  `gorm:"column:hash" json:"hash"`                                          // 哈希值：由原值生成，通过sha256算法 原值：服务器种子+比特币区块哈希+轮数
	CrashMulti    float64 `gorm:"column:crash_multi;type:decimal(10,2)" json:"crash_multi"`         // 爆炸倍数
	CrashDuration int64   `gorm:"column:crash_duration" json:"crash_duration"`                      // 爆炸持续时间
	WaitingTime   int64   `gorm:"column:waiting_time" json:"waiting_time"`                          // 等待时间
	Settled       uint8   `gorm:"column:settled" json:"settled"`                                    // 是否已结算
}

func (c *CrashGameRound) TableName() string {
	return "crash_game_round"
}

type CrashGameOrder struct {
	BaseModel
	UID              uint    `gorm:"column:uid;uniqueIndex:idx_uid_round_id_bet_index" json:"uid"`                                   // 用户ID 注意索引的名称和排序字段
	Name             string  `gorm:"column:name;size:64" json:"name"`                                                                // 用户名
	RoundID          uint64  `gorm:"column:round_id;size:35;uniqueIndex:idx_uid_round_id_bet_index" json:"round_id"`                 // 轮数：每进行一局，轮数+1
	BetIndex         int     `gorm:"column:bet_index;default:0;uniqueIndex:idx_uid_round_id_bet_index" json:"bet_index"`             // 投注索引
	AutoEscapeHeight float64 `gorm:"column:auto_escape_height;default:0;type:decimal(10,2)" json:"auto_escape_height"`               // 自动逃跑高度
	EscapeHeight     float64 `gorm:"column:escape_height;default:0;type:decimal(10,2);index:idx_escape_height" json:"escape_height"` // 逃跑高度
	EscapeTime       int64   `gorm:"column:escape_time" json:"escape_time"`                                                          // 逃跑时间
	Rate             uint8   `gorm:"column:rate;default:0" json:"-"`                                                                 // 抽水比例
	BetTime          int64   `gorm:"column:bet_time" json:"bet_time"`                                                                // 投注时间
	BetAmount        float64 `gorm:"column:bet_amount;default:0;type:decimal(10,2)" json:"bet_amount"`                               // 投注金额
	Delivery         float64 `gorm:"column:delivery;default:0;type:decimal(10,2)" json:"delivery"`                                   // 下注减抽水
	Fee              float64 `gorm:"column:fee;default:0;type:decimal(10,2)" json:"fee"`                                             // 抽水
	RewardAmount     float64 `gorm:"column:reward_amount;default:0;type:decimal(20,2)" json:"reward_amount"`                         // 中奖金额
	PromoterCode     int     `gorm:"column:pc;default:0" json:"-"`
	Status           uint8   `gorm:"column:status;default:0" json:"status"` // 0 未结算 1 已结算 2 已取消
	EndTime          int64   `gorm:"end_time" json:"end_time"`              // 单的应完成结算时间

	OrderID string `gorm:"-" json:"order_id"` // 临时orderID，不存储到数据库
}

func (o *CrashGameOrder) TableName() string {
	return "crash_game_order"
}

func (o *CrashGameOrder) CalculateFee() { //抽水处理
	decimalBetAmount := decimal.NewFromFloat(o.BetAmount)
	decimalRate := decimal.NewFromFloat(float64(o.Rate) / 1000)
	decimalFee := decimalBetAmount.Mul(decimalRate)
	decimalDelivery := decimalBetAmount.Sub(decimalFee)
	fee := math.MustParsePrecFloat64(decimalFee.InexactFloat64(), 3)
	delivery := math.MustParsePrecFloat64(decimalDelivery.InexactFloat64(), 3)
	o.Delivery = delivery
	o.Fee = fee
}

type CrashAutoBet struct {
	BaseModel
	UID              uint    `gorm:"column:uid;uniqueIndex:idx_uid" json:"uid"`
	BetAmount        float64 `gorm:"column:bet_amount;default:0;type:decimal(10,2)" json:"bet_amount"`                 // 投注金额
	AutoEscapeHeight float64 `gorm:"column:auto_escape_height;default:0;type:decimal(10,2)" json:"auto_escape_height"` // 自动逃跑高度
	AutoBetCount     uint64  `gorm:"column:auto_bet_count;default:0"  json:"auto_bet_count"`                           // 自动下注次数，大于0固定次数，为0无穷次
	IsInfinite       uint8   `gorm:"column:is_infinite;default:0"  json:"is_infinite"`                                 // 0 固定次数 1 无穷次
	Status           uint8   `gorm:"column:status;default:0" json:"status"`                                            // 0 未生效 1 已生效
}

func (c *CrashAutoBet) TableName() string {
	return "crash_auto_bet"
}

// -------------------------------- request/response -------------------------------

type GetCrashGameRoundReq struct {
	RoundID uint64 `json:"round_id"` // 回合id，默认为空获取最新回合
}

type GetCrashGameRoundRsp struct {
	RoundID    uint64  `json:"round_id"`    // 回合id
	Status     string  `json:"status"`      // 状态: preparing:准备,上一局结算时间5.5s betting:下注7s flying:飞行,具体时间根据爆炸倍数计算 settling:结算 completed:已完成
	ServerSeed string  `json:"server_seed"` // 服务器种子
	Hash       string  `json:"hash"`        // 哈希值
	OpenHash   string  `json:"open_hash"`   // 提前公布的哈希值：由哈希值，通过sha256算法生成，会提前公布给用户
	CrashMulti float64 `json:"crash_multi"` // 爆炸倍数

	CurrentTime       int64 `json:"current_time"`        // 当前时间
	CurrentStatusTime int64 `json:"current_status_time"` // 当前状态时间
	Settled           uint8 `json:"settled"`             // 是否已结算

	UltimateEscaper string  `json:"ultimate_escaper"` // 最终逃跑者 用户昵称，已结算的回合才有值
	ExtremeAltitude float64 `json:"extreme_altitude"` // 极限高度 用户昵称，已结算的回合才有值

	NotifyType string `json:"notify_type"` // 类型  order:下注 round:回合
}

type CrashGameOrderNotify struct {
	UID          uint    `json:"uid"`           // 用户ID
	Name         string  `json:"name"`          // 用户名
	RoundID      uint64  `json:"round_id"`      // 回合id
	BetIndex     int     `json:"bet_index"`     // 投注索引
	BetAmount    float64 `json:"bet_amount"`    // 投注金额
	BetTime      int64   `json:"bet_time"`      // 投注时间
	EscapeHeight float64 `json:"escape_height"` // 逃跑高度
	EscapeTime   int64   `json:"escape_time"`   // 逃跑时间
	RewardAmount float64 `json:"reward_amount"` // 中奖金额
	Status       string  `json:"status"`        // 状态  bet:下注 escape:逃跑
	NotifyType   string  `json:"notify_type"`   // 类型  order:下注 round:回合
}

type GetCrashGameRoundListReq struct {
}

type GetCrashGameRoundListRsp struct {
	List                  []*GetCrashGameRoundRsp `json:"list"`                    // 回合
	AverageEscapeAltitude float64                 `json:"average_escape_altitude"` // 平均逃跑高度
}

type GetCrashGameRoundOrderListReq struct {
	RoundID uint64 `json:"round_id"` // 回合id，默认为空获取最新回合
}

type PlaceCrashGameBetReq struct {
	UID uint `json:"-"`
	// RoundID          uint64  `json:"round_id" binding:"required,gt=0"` // 回合id
	BetIndex         int     `json:"bet_index"` // 投注索引
	BetAmount        float64 `json:"bet_amount" binding:"required,gt=0"`
	AutoEscapeHeight float64 `json:"auto_escape_height"`
}

type CancelCrashGameBetReq struct {
	UID uint `json:"-"`
	// RoundID uint64 `json:"round_id" binding:"required,gt=0"` // 回合id
	BetIndex int `json:"bet_index"` // 投注索引
}

type EscapeCrashGameBetReq struct {
	UID uint `json:"-"`
	// RoundID      uint64  `json:"round_id" binding:"required,gt=0"` // 回合id
	BetIndex     int     `json:"bet_index"` // 投注索引
	EscapeHeight float64 `json:"escape_height" binding:"required,gt=0"`
	EscapeTime   int64   `json:"escape_time" binding:"required,gt=0"`
}

type GetCrashAutoBetReq struct {
	UID uint `json:"-"`
}

type PlaceCrashAutoBetReq struct {
	UID              uint    `json:"-"`
	BetAmount        float64 `json:"bet_amount"` // 自动下注次数，大于0固定次数，为0无穷次
	AutoEscapeHeight float64 `json:"auto_escape_height" binding:"required,gt=0"`
	AutoBetCount     uint64  `json:"auto_bet_count"`
}

type CancelCrashAutoBetReq struct {
	UID uint `json:"-"`
}

type GetUserCrashGameOrderReq struct {
	UID     uint   `json:"-"`
	RoundID uint64 `json:"round_id"` // 回合id，默认为空获取最新回合
}

type GetUserCrashGameOrderListReq struct {
	UID uint `json:"-"`
}

type GetUserCrashGameOrderListRsp struct {
	List              []*CrashGameOrder `json:"list"`
	TotalBetAmount    float64           `json:"total_bet_amount"`    // 总下注金额
	TotalRewardAmount float64           `json:"total_reward_amount"` // 总中奖金额
}
