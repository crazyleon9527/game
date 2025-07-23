package entities

import (
	"rk-api/internal/app/errors"
	"rk-api/pkg/math"

	"github.com/shopspring/decimal"
)

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

// WingoRoomSetting 设置表
type WingoRoomSetting struct {
	BaseModel           `json:"-"`
	BetType             uint8  `gorm:"column:bet_type" json:"betType"`                          // 房间类型
	RoundInterval       uint   `gorm:"column:round_interval" json:"roundInterval"`              //回合时间
	BettingInterval     uint   `gorm:"column:betting_interval" json:"bettingInterval"`          // 投注时长
	StopBettingInterval uint   `gorm:"column:stop_betting_interval" json:"stopBettingInterval"` // 停止投注时长
	SettleInterval      uint   `gorm:"column:settle_interval" json:"settleInterval"`            // 结算投注时长
	Rate                uint8  `gorm:"column:rate" json:"rate"`                                 // 抽水比例
	Currency            string `gorm:"column:currency;size:20" json:"currency"`                 // 货币
	Status              uint8  `gorm:"column:status" json:"status"`                             // 状态
}

// WingoPeriod 期数表
type WingoPeriod struct {
	BaseModel
	BetType      uint8   `gorm:"column:bet_type;default:0" json:"betType"`                   // 房间类型
	Rate         uint8   `json:"rate"`                                                       // 抽水
	PeriodID     string  `gorm:"column:period;size:35;" json:"periodID"`                     // 期数
	PeriodDate   string  `gorm:"column:period_date;size:8;default:0" json:"-"`               // 日期
	PeriodIndex  uint    `gorm:"column:period_index;default:0" json:"periodIndex"`           // 期数序号
	PresetNumber int8    `gorm:"column:preset_number;" json:"-"`                             // 预设的数字
	Number       int8    `gorm:"column:number;default:0" json:"number"`                      // 开的数字
	PlayerCount  uint    `gorm:"column:player_count;default:0" json:"-"`                     // 玩家数目
	OrderCount   uint    `json:"-"`                                                          // 订单数目
	BetAmount    float64 `gorm:"column:bet_amount;default:0;type:decimal(10,2)" json:"-"`    // 投注金额
	RewardAmount float64 `gorm:"column:reward_amount;default:0;type:decimal(10,2)" json:"-"` // 中奖金额
	Price        float64 `gorm:"column:price;default:0;type:decimal(10,2)" json:"price"`     // 模拟计算的爆奖额
	Fee          float64 `gorm:"column:fee;default:0;type:decimal(10,2)" json:"-"`           // 手续费
	Profit       float64 `gorm:"column:profit;default:0;type:decimal(10,2)" json:"-"`        // 盈利
	StartTime    int64   `json:"startTime"`                                                  //开始时间
	EndTime      int64   `json:"endTime"`                                                    //结束时间
	Status       uint8   `gorm:"column:status;default:0" json:"status"`                      // 状态
}

// column:status;default:0

// // 自定义JSON序列化功能
// func (wp WingoPeriod) MarshalJSON() ([]byte, error) {
// 	type Alias WingoPeriod
// 	return json.Marshal(&struct {
// 		ID uint `json:"id"`
// 		*Alias
// 	}{
// 		ID:    wp.ID,
// 		Alias: (*Alias)(&wp),
// 	})
// }

// WingoOrder 订单表
type WingoOrder struct {
	// BaseModel
	BaseModel
	UID          uint    `gorm:"column:uid;index:idx_uid_bet_type"`                                     // 用户ID 注意索引的名称和排序字段
	BetType      uint8   `gorm:"column:bet_type;index:idx_uid_bet_type"  json:"betType"`                // 房间类型  同时作为复合索引的一部分
	PeriodID     string  `gorm:"column:period;size:35;" json:"periodID"`                                // 期数
	TicketNumber uint8   `gorm:"column:ticket_number;default:0" json:"ticketNumber"`                    // 彩票ID 选的数字
	Number       int8    `gorm:"column:number;default:0" json:"number"`                                 // 实际开奖数字
	Rate         uint8   `json:"-"`                                                                     // 抽水比例
	BetTime      int64   `json:"betTime"`                                                               //投注时间
	BetAmount    float64 `gorm:"column:bet_amount;default:0;type:decimal(10,2)" json:"betAmount"`       // 投注金额
	Fee          float64 `gorm:"column:fee;default:0;type:decimal(10,2)" json:"fee"`                    // 抽水
	Delivery     float64 `gorm:"column:delivery;default:0;type:decimal(10,2)" json:"delivery"`          // 下注减抽水
	RewardAmount float64 `gorm:"column:reward_amount;default:0;type:decimal(10,2)" json:"rewardAmount"` // 中奖金额
	Price        float64 `gorm:"column:price;default:0;type:decimal(10,2)" json:"price"`                // 模拟计算的爆奖额
	Balance      float64 `gorm:"column:balance;default:0;type:decimal(10,2)" json:"balance"`            // 下注后余额
	Status       uint8   `gorm:"column:status;default:0" json:"status"`                                 // 是否已经结算
	PromoterCode int     `gorm:"column:pc;default:0" json:"-"`
	FinishTime   int64   `gorm:"finish_time" json:"finishTime"` //单的应完成结算时间
	Username     string  `gorm:"-"  json:"username"`
	Color        uint8   `gorm:"-"  json:"color"`
}

func (o *WingoOrder) CalculateFee() { //抽水处理
	decimalBetAmount := decimal.NewFromFloat(o.BetAmount)
	decimalRate := decimal.NewFromFloat(float64(o.Rate) / 1000)
	decimalFee := decimalBetAmount.Mul(decimalRate)
	decimalDelivery := decimalBetAmount.Sub(decimalFee)
	fee := math.MustParsePrecFloat64(decimalFee.InexactFloat64(), 3)
	delivery := math.MustParsePrecFloat64(decimalDelivery.InexactFloat64(), 3)

	o.Fee = fee
	o.Delivery = delivery
}

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////
type WingoOrderReq struct {
	UID          uint   `json:"uid"  `                         // 平台用户ID
	PeriodID     string `json:"periodID"  binding:"required"`  // 期数ID，omitempty表示若字段为空，则不显示在JSON中
	TicketNumber uint   `json:"ticketNumber"  `                // 彩票数字，注意这里JSON标记要与struct中的字段名称匹配
	BetAmount    uint   `json:"betAmount"  binding:"required"` // 投注金额
	BetType      uint8  `json:"betType"  binding:"required"`   // 房间类型
}

func (o *WingoOrderReq) CheckTicketInvalid() error {
	if (o.TicketNumber <= 9) || o.TicketNumber == 10 || o.TicketNumber == 20 || o.TicketNumber == 30 {
		return nil
	}
	return errors.WithCode(errors.InvalidParam)
}

type WingoOrderHistoryReq struct {
	Paginator
	UID     uint
	BetType uint8 `json:"betType" binding:"required"` // 房间类型，可以为空
}

type GetPeriodHistoryListReq struct {
	Paginator
	BetType uint8 `json:"betType" binding:"required"` // 房间类型，可以为空
}

type GetRoomReq struct {
	BetType uint8 `json:"betType" binding:"required"` // 房间类型，必填项
}

type StateSyncReq struct {
	PeriodID string `json:"periodID" binding:"required"` // 期数ID，
	BetType  uint8  `json:"betType" binding:"required"`  // 房间类型，
	State    string `json:"state" binding:"required"`    // 状态，必填项
}

type OrderReq struct {
	PromoterCode *int `json:"pc,omitempty" binding:"omitempty"` // 分销码，可以为空
	Start        int  `json:"start" binding:"gte=0"`            // 起始，大于等于0
	Num          int  `json:"num" binding:"gt=0"`               // 数量，大于0
	OrderBy      int  `json:"order" binding:"omitempty"`        // 排序方式，可以为空
	BetType      uint `json:"betType" binding:"omitempty"`      // 房间类型，可以为空
}

type UpdatePeriodReq struct {
	PeriodID string `json:"periodID" binding:"required"` // 期数ID，可以为空
	BetType  uint8  `json:"betType" binding:"required"`  // 房间类型，可以为空
	Number   int    `json:"number" `                     // 数字，可以为空
}

type WingoRoomResp struct {
	ID          uint              `json:"id"`          // ID
	Setting     *WingoRoomSetting `json:"setting"`     // 设置
	PeriodID    string            `json:"periodID"`    // 期数ID
	StateSTime  int64             `json:"stateSTime"`  // 状态时间
	RoundSTime  int64             `json:"roundSTime"`  // 回合时间
	NowSTime    int64             `json:"nowSTime"`    // 当前时间
	PlayerCount int               `json:"playerCount"` // 玩家数目
	State       string            `json:"state"`       // 状态
	PeriodIndex uint              `json:"periodIndex"` // 期数索引
}

type StateResp struct {
	ID          uint   `json:"id"`          // ID
	PeriodID    string `json:"periodID"`    // 期数ID
	StateSTime  int64  `json:"stateSTime"`  // 状态时间
	RoundSTime  int64  `json:"roundSTime"`  // 回合时间
	NowSTime    int64  `json:"nowSTime"`    // 当前时间
	PlayerCount int    `json:"playerCount"` // 玩家数目
	State       string `json:"state"`       // 状态
	PeriodIndex uint   `json:"periodIndex"` // 期数索引
	// PlayerCash  *float64 `json:"playerCash"`  // 玩家当前金额
}

type GetPeriodListReq struct {
	Paginator
	BetType uint8 `json:"betType" ` // 房间类型，
}

type GetPeriodBetInfoReq struct {
	BetType       uint8 `json:"betType"`
	PromotionCode uint  `json:"pc"`
}

type GetPeriodReq struct {
	BetType uint8 `json:"betType"`
}

type UpdateRoomLimitReq struct {
	InPcRoomLimitState bool `json:"inPcRoomLimitState"` //是否处于pc
}

// 给后台使用的periodInfo  兼容旧的
type PeriodInfo struct {
	PeriodID     string `json:"periodId"`
	PresetNumber int8   `json:"presetNumber"`
	DefineNumber int8   `json:"defineNumber"`

	Status    uint8 `json:"status"`
	CountDown int64 `json:"countDown"` //倒计时截至时间
	Time      int64 `json:"time"`      //表示现在服务器时间 兼容旧的

	InPcRoomLimitState bool `json:"inPcRoomLimitState"` //是否处于pc
}

type PeriodResult struct {
	Number      int8 `json:"n"`
	PeriodIndex uint `json:"i"`
}

type TrendInfo struct {
	RedCount    int `json:"reds"`
	VioletCount int `json:"violets"`
	GreenCount  int `json:"greens"`
	// Numbers      []int8          `json:"numbers"`
	// Indexs       []uint          `json:"indexs"`
	Results []*PeriodResult `json:"results"`
}

type WingoTrendReq struct {
	BetType uint8 `json:"betType"  binding:"required"` // 房间类型
}

// Assuming Paginator and WingoRoomSetting have their own fields and validation,
// ommitted here for clarity.

// {\"incID\":160,\"periodId\":\"20220328160\",\"status\":0,\"startTime\":1648498264138,\"countDown\":1648498564138,\"presetNumber\":4,\"timeStamp\":\"202203281711\"}"
