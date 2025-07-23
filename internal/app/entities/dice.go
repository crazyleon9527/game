package entities

import (
	"rk-api/pkg/math"

	"github.com/shopspring/decimal"
)

type DiceGameOrder struct {
	BaseModel
	UID          uint    `gorm:"column:uid;index:idx_uid;uniqueIndex:idx_uid_round_id" json:"uid"`                        // 用户ID
	RoundID      uint64  `gorm:"column:round_id;size:35;index:idx_round_id;uniqueIndex:idx_uid_round_id" json:"round_id"` // 轮数
	ClientSeed   string  `gorm:"column:client_seed;size:64" json:"client_seed"`                                           // 客户端种子
	ServerSeed   string  `gorm:"column:server_seed;size:64" json:"server_seed"`                                           // 服务端种子
	Target       float64 `gorm:"column:target;default:0;type:decimal(10,2)" json:"target"`                                // 目标值
	Result       float64 `gorm:"column:result;default:0;type:decimal(10,2)" json:"result"`                                // 结果值
	IsAbove      int     `gorm:"column:is_above;default:0" json:"is_above"`                                               // 是否大于 1:大于 0:小于
	Multiple     float64 `gorm:"column:multiple;default:0;type:decimal(10,4)" json:"multiple"`                            // 倍数
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

func (o *DiceGameOrder) TableName() string {
	return "dice_game_order"
}

func (o *DiceGameOrder) CalculateFee() { //抽水处理
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

type DiceGameState struct {
	RoundID      uint64  `json:"round_id"`      // 轮数
	ClientSeed   string  `json:"client_seed"`   // 客户端种子
	ServerSeed   string  `json:"server_seed"`   // 服务端种子
	OpenHash     string  `json:"open_hash"`     // 提前公布的hash
	Target       float64 `json:"target"`        // 目标值
	Result       float64 `json:"result"`        // 结果值
	IsAbove      int     `json:"is_above"`      // 是否大于 1:大于 0:小于
	Multiple     float64 `json:"multiple"`      // 倍数
	BetTime      int64   `json:"bet_time"`      // 投注时间
	BetAmount    float64 `json:"bet_amount"`    // 投注金额
	RewardAmount float64 `json:"reward_amount"` // 中奖金额
	Settled      uint8   `json:"settled"`       // 是否已结算
	EndTime      int64   `json:"end_time"`      // 单的完成结算时间
}

type DiceGameGetStateReq struct {
	UID uint `json:"-"`
}

type DiceGamePlaceBetReq struct {
	UID       uint    `json:"-"`
	BetAmount float64 `json:"bet_amount" binding:"required,gt=0"`
	Target    float64 `json:"target" binding:"required,gt=0"`
	IsAbove   int     `json:"is_above"` // 是否大于 1:大于 0:小于
}

type DiceGameOpenPositionReq struct {
	UID uint `json:"-"`
}

type DiceGameCashoutReq struct {
	UID uint `json:"-"`
}

type DiceGameChangeSeedReq struct {
	UID        uint   `json:"-"`
	ClientSeed string `json:"client_seed"` // 客户端种子
}

type DiceGameChangeSeedRsp struct {
	ClientSeed string `json:"client_seed"` // 客户端种子
	OpenHash   string `json:"open_hash"`   // 提前公布的hash
}

type DiceGameGetOrderListReq struct {
	UID uint `json:"-"`
}
