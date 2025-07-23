package entities

import (
	"errors"
	"rk-api/pkg/math"

	"github.com/shopspring/decimal"
)

// --------------------------------- 数据库表 ---------------------------------

// 游戏回合接口
type IHashGameRound interface {
	TableName() string

	GetRoundID() string
	GetBlockHeight() uint64
	GetStatus() string
	GetHash() string
	GetEndTime() int64
	GetSettled() uint8
	GetResult() string
}

type BaseHashGameRound struct {
	BaseModel
	RoundID     string `gorm:"column:roundId;size:35;" json:"roundId"`          // 期数
	BlockHeight uint64 `gorm:"column:blockHeight;default:0" json:"blockHeight"` // 区块
	Status      string `gorm:"column:status;default:0" json:"status"`           // 状态
	Hash        string `gorm:"hash" json:"hash"`                                // 区块hash
	EndTime     int64  `gorm:"end_time" json:"endTime"`                         // 结束时间
	Settled     uint8  `gorm:"settled" json:"settled"`                          // 是否已结算
	Result      string `gorm:"result" json:"result"`                            // 结果
}

func (t *BaseHashGameRound) GetRoundID() string     { return t.RoundID }
func (t *BaseHashGameRound) GetBlockHeight() uint64 { return t.BlockHeight }
func (t *BaseHashGameRound) GetStatus() string      { return t.Status }
func (t *BaseHashGameRound) GetHash() string        { return t.Hash }
func (t *BaseHashGameRound) GetEndTime() int64      { return t.EndTime }
func (t *BaseHashGameRound) GetSettled() uint8      { return t.Settled }
func (t *BaseHashGameRound) GetResult() string      { return t.Result }

// 游戏回合 单双
type HashSDGameRound struct {
	*BaseHashGameRound
}

func (t *HashSDGameRound) TableName() string {
	return "hash_sd_game_round"
}

// 游戏回合订单接口
type IHashGameOrder interface {
	TableName() string

	GetID() uint
	GetUID() uint
	GetBetType() uint8
	GetRoundID() string
	GetRate() uint8
	GetBetTime() int64
	GetBetAmount() float64
	GetDelivery() float64
	GetFee() float64
	GetRewardAmount() float64
	SetRewardAmount(float64)
	GetStatus() uint8
	GetEndTime() int64
	SetEndTime(int64)
	GetPromoterCode() int
	SetPromoterCode(int)
	GetPrediction() uint8
	GetOrderID() string

	CalculateFee()
}

type BaseHashGameOrder struct {
	BaseModel
	UID          uint    `gorm:"column:uid;index:idx_uid_bet_type"`                                     // 用户ID 注意索引的名称和排序字段
	BetType      uint8   `gorm:"column:bet_type;index:idx_uid_bet_type"  json:"betType"`                // 房间类型  同时作为复合索引的一部分 1 单双 2 大小 3 bullbull 4 banker player tie 5 lucky
	RoundID      string  `gorm:"column:roundId;size:35;" json:"roundId"`                                // 期数
	Rate         uint8   `json:"-"`                                                                     // 抽水比例
	BetTime      int64   `json:"betTime"`                                                               // 投注时间
	BetAmount    float64 `gorm:"column:bet_amount;default:0;type:decimal(10,2)" json:"betAmount"`       // 投注金额
	Delivery     float64 `gorm:"column:delivery;default:0;type:decimal(10,2)" json:"delivery"`          // 下注减抽水
	Fee          float64 `gorm:"column:fee;default:0;type:decimal(10,2)" json:"fee"`                    // 抽水
	RewardAmount float64 `gorm:"column:reward_amount;default:0;type:decimal(10,2)" json:"rewardAmount"` // 中奖金额
	Status       uint8   `gorm:"column:status;default:0" json:"status"`                                 // 是否已经结算
	PromoterCode int     `gorm:"column:pc;default:0" json:"-"`
	EndTime      int64   `gorm:"end_time" json:"endTime"`      // 单的应完成结算时间
	Prediction   uint8   `gorm:"prediction" json:"prediction"` // 单双 1:单 2:双  大小 1:小 2:大  bullbull 1:庄牛牛 2:闲牛牛 3:庄牛九 4:闲牛九 5:庄赢 6:闲赢  lucky 1:不中 2:中  banker player tie 1:庄 2:闲 3:和
	OrderID      string  `gorm:"-" json:"orderID"`             // 临时orderID，不存储到数据库
}

func (r *BaseHashGameOrder) GetID() uint               { return r.ID }
func (r *BaseHashGameOrder) GetUID() uint              { return r.UID }
func (r *BaseHashGameOrder) GetBetType() uint8         { return r.BetType }
func (r *BaseHashGameOrder) GetRoundID() string        { return r.RoundID }
func (r *BaseHashGameOrder) GetRate() uint8            { return r.Rate }
func (r *BaseHashGameOrder) GetBetTime() int64         { return r.BetTime }
func (r *BaseHashGameOrder) GetBetAmount() float64     { return r.BetAmount }
func (r *BaseHashGameOrder) SetRewardAmount(v float64) { r.RewardAmount = v }
func (r *BaseHashGameOrder) GetDelivery() float64      { return r.Delivery }
func (r *BaseHashGameOrder) GetFee() float64           { return r.Fee }
func (r *BaseHashGameOrder) GetRewardAmount() float64  { return r.RewardAmount }
func (r *BaseHashGameOrder) GetStatus() uint8          { return r.Status }
func (r *BaseHashGameOrder) GetEndTime() int64         { return r.EndTime }
func (r *BaseHashGameOrder) SetEndTime(v int64)        { r.EndTime = v }
func (r *BaseHashGameOrder) GetPromoterCode() int      { return r.PromoterCode }
func (r *BaseHashGameOrder) SetPromoterCode(v int)     { r.PromoterCode = v }
func (r *BaseHashGameOrder) GetPrediction() uint8      { return r.Prediction }
func (r *BaseHashGameOrder) GetOrderID() string        { return r.OrderID }

func (o *BaseHashGameOrder) CalculateFee() { //抽水处理
	decimalBetAmount := decimal.NewFromFloat(o.BetAmount)
	decimalRate := decimal.NewFromFloat(float64(o.Rate) / 1000)
	decimalFee := decimalBetAmount.Mul(decimalRate)
	decimalDelivery := decimalBetAmount.Sub(decimalFee)
	fee := math.MustParsePrecFloat64(decimalFee.InexactFloat64(), 3)
	delivery := math.MustParsePrecFloat64(decimalDelivery.InexactFloat64(), 3)
	o.Delivery = delivery
	o.Fee = fee
}

// 游戏回合订单 单双
type HashSDGameOrder struct {
	*BaseHashGameOrder
}

func (o *HashSDGameOrder) TableName() string {
	return "hash_sd_game_order"
}

// -------------------------------------- 接口 --------------------------------------

// 通用投注接口
type IHashBetRequest interface {
	GetUID() uint
	GetBetType() uint8
	GetBetAmount() float64
	GetPrediction() uint8

	Validate() error
}

// 基础投注请求
type BaseHashBetRequest struct {
	UID        uint    `json:"-"`
	BetType    uint8   `json:"bet_type" binding:"required"` // 房间类型 1 初级 2 中级 3 高级
	BetAmount  float64 `json:"bet_amount" binding:"required,gt=0"`
	Prediction uint8   `json:"prediction"` // 1:单双 (1:单 2:双)  2:大小 (1:小 2:大)  3:bullbull (1:庄牛牛 2:闲牛牛 3:庄牛九 4:闲牛九 5:庄赢 6:闲赢)  4:lucky (1:不中 2:中)  5:banker player tie (1:庄 2:闲 3:和)
}

func (r *BaseHashBetRequest) GetUID() uint          { return r.UID }
func (r *BaseHashBetRequest) GetBetType() uint8     { return r.BetType }
func (r *BaseHashBetRequest) GetBetAmount() float64 { return r.BetAmount }
func (r *BaseHashBetRequest) GetPrediction() uint8  { return r.Prediction }
func (r *BaseHashBetRequest) Validate() error       { return nil }

// 单双玩法请求
type HashSDBetRequest struct {
	BaseHashBetRequest
}

// 牛牛玩法请求
type HashBullBullBetRequest struct {
	BaseHashBetRequest
	Position  string `json:"position" binding:"required,oneof=player banker tie"` // 押注位置
	CardCount int    `json:"card_count" binding:"required,gte=1,lte=5"`           // 押注牌数
}

func (r *HashBullBullBetRequest) Validate() error {
	if r.CardCount < 1 || r.CardCount > 5 {
		return errors.New("invalid card count")
	}
	return nil
}

// -------------------------------------- 其它 --------------------------------------

type FairCheckReq struct {
	Game       string            `json:"game"`        // 游戏名称 "Crash","Mine"
	ClientSeed string            `json:"client_seed"` // 客户端随机种子
	ServerSeed string            `json:"server_seed"` // 服务端随机种子
	OpenHash   string            `json:"open_hash"`   // 提前公布的hash
	BlockHash  string            `json:"block_hash"`  // 区块hash
	Ext        map[string]string `json:"ext"`         // 扩展参数
}

type FairCheckRsp struct {
	Result     float64 `json:"result"`
	ResultJson string  `json:"result_json"`
}
