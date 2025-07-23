package entities

import (
	"rk-api/pkg/math"

	"github.com/shopspring/decimal"
)

type MineGameOrder struct {
	BaseModel
	UID          uint    `gorm:"column:uid;index:idx_uid;uniqueIndex:idx_uid_round_id" json:"uid"`                        // 用户ID
	RoundID      uint64  `gorm:"column:round_id;size:35;index:idx_round_id;uniqueIndex:idx_uid_round_id" json:"round_id"` // 轮数
	Status       string  `gorm:"column:status" json:"status"`                                                             // 状态: preparing:准备 playing:进行中 gameover:游戏结束
	ClientSeed   string  `gorm:"column:client_seed;size:64" json:"client_seed"`                                           // 客户端种子
	ServerSeed   string  `gorm:"column:server_seed;size:64" json:"server_seed"`                                           // 服务端种子
	MineCount    int     `gorm:"column:mine_count;default:0" json:"mine_count"`                                           // 地雷个数
	DiamondLeft  int     `gorm:"column:diamond_left;default:0" json:"diamond_left"`                                       // 剩余钻石个数
	MinePosition string  `gorm:"column:mine_position;size:128" json:"mine_position"`                                      // 地雷位置json 0-24  [17,5,2]
	OpenPosition string  `gorm:"column:open_position;size:1024" json:"open_position"`                                     // 开启的位置json 0-24  [{"position": 17,"multiple":1.13}]
	Multiple     float64 `gorm:"column:multiple;default:0;type:decimal(10,2)" json:"multiple"`                            // 倍数
	Rate         uint8   `gorm:"column:rate;default:0" json:"-"`                                                          // 抽水比例
	BetTime      int64   `gorm:"column:bet_time" json:"bet_time"`                                                         // 投注时间
	BetAmount    float64 `gorm:"column:bet_amount;default:0;type:decimal(10,2)" json:"bet_amount"`                        // 投注金额
	Delivery     float64 `gorm:"column:delivery;default:0;type:decimal(10,2)" json:"delivery"`                            // 下注减抽水
	Fee          float64 `gorm:"column:fee;default:0;type:decimal(10,2)" json:"fee"`                                      // 抽水
	RewardAmount float64 `gorm:"column:reward_amount;default:0;type:decimal(20,2)" json:"reward_amount"`                  // 中奖金额
	PromoterCode int     `gorm:"column:pc;default:0" json:"-"`
	Settled      uint8   `gorm:"column:settled;index:idx_settled" json:"settled"` // 是否已结算
	EndTime      int64   `gorm:"end_time" json:"end_time"`                        // 单的应完成结算时间
}

func (o *MineGameOrder) TableName() string {
	return "mine_game_order"
}

func (o *MineGameOrder) CalculateFee() { //抽水处理
	decimalBetAmount := decimal.NewFromFloat(o.BetAmount)
	decimalRate := decimal.NewFromFloat(float64(o.Rate) / 1000)
	decimalFee := decimalBetAmount.Mul(decimalRate)
	decimalDelivery := decimalBetAmount.Sub(decimalFee)
	fee := math.MustParsePrecFloat64(decimalFee.InexactFloat64(), 3)
	delivery := math.MustParsePrecFloat64(decimalDelivery.InexactFloat64(), 3)
	o.Delivery = delivery
	o.Fee = fee
}

// ------------------------------------------------ 请求/响应 -----------------------------------------------

type MineGamePosition struct {
	Position int     `json:"position"`
	Multiple float64 `json:"multiple"`
}

type MineGameState struct {
	RoundID      uint64              `json:"round_id"`      // 轮数
	Status       string              `json:"status"`        // 状态 : preparing:准备 playing:进行中 gameover:游戏结束
	ClientSeed   string              `json:"client_seed"`   // 客户端种子
	ServerSeed   string              `json:"server_seed"`   // 服务端种子
	OpenHash     string              `json:"open_hash"`     // 提前公布的hash
	MineCount    int                 `json:"mine_count"`    // 地雷个数
	DiamondLeft  int                 `json:"diamond_left"`  // 剩余钻石个数
	MinePosition []int               `json:"mine_position"` // 地雷位置json 0-24  [17,5,2]
	OpenPosition []*MineGamePosition `json:"open_position"` // 开启的位置json 0-24  [{"position": 17,"multiple":1.13}]
	Multiple     float64             `json:"multiple"`      // 倍数
	BetTime      int64               `json:"bet_time"`      // 投注时间
	BetAmount    float64             `json:"bet_amount"`    // 投注金额
	RewardAmount float64             `json:"reward_amount"` // 中奖金额
	Settled      uint8               `json:"settled"`       // 是否已结算
	EndTime      int64               `json:"end_time"`      // 单的完成结算时间
}

type MineGameGetStateReq struct {
	UID uint `json:"-"`
}

type MineGamePlaceBetReq struct {
	UID       uint    `json:"-"`
	BetAmount float64 `json:"bet_amount" binding:"required,gt=0"`
	MineCount int     `json:"mine_count" binding:"required,gt=0"`
}

type MineGameOpenPositionReq struct {
	UID          uint `json:"-"`
	OpenPosition int  `json:"open_position"`
}

type MineGameCashoutReq struct {
	UID uint `json:"-"`
}

type MineGameChangeSeedReq struct {
	UID        uint   `json:"-"`
	ClientSeed string `json:"client_seed"` // 客户端种子
}

type MineGameChangeSeedRsp struct {
	ClientSeed string `json:"client_seed"` // 客户端种子
	OpenHash   string `json:"open_hash"`   // 提前公布的hash
}

type MineGameGetOrderListReq struct {
	UID uint `json:"-"`
}
