package entities

import (
	"regexp"
	"rk-api/internal/app/errors"
	"rk-api/pkg/math"
	"strings"

	"github.com/shopspring/decimal"
)

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

// NineRoomSetting 设置表
type NineRoomSetting struct {
	BaseModel     `json:"-"`
	BetType       uint8 `gorm:"column:bet_type" json:"betType"`             // 房间类型
	RoundInterval uint  `gorm:"column:round_interval" json:"roundInterval"` //回合时间

	BettingInterval     uint   `gorm:"column:betting_interval" json:"bettingInterval"`          // 投注时长
	StopBettingInterval uint   `gorm:"column:stop_betting_interval" json:"stopBettingInterval"` // 停止投注时长
	SettleInterval      uint   `gorm:"column:settle_interval" json:"settleInterval"`            // 结算投注时长
	Rate                uint8  `gorm:"column:rate" json:"rate"`                                 // 抽水比例
	Currency            string `gorm:"column:currency;size:20" json:"currency"`                 // 货币
	Status              uint8  `gorm:"column:status" json:"status"`                             // 状态
}

// NinePeriod 期数表
type NinePeriod struct {
	BaseModel
	BetType      uint8   `gorm:"column:bet_type;default:0" json:"betType"`                   // 房间类型
	Rate         uint8   `json:"rate"`                                                       // 抽水
	PeriodID     string  `gorm:"column:period;size:35;" json:"periodID"`                     // 期数
	PeriodDate   string  `gorm:"column:period_date;size:8;default:0" json:"-"`               // 日期
	PeriodIndex  uint    `gorm:"column:period_index;default:0" json:"periodIndex"`           // 期数序号
	PresetNumber int8    `gorm:"column:preset_number;" json:"-"`                             // 预设的数字
	Number       int8    `gorm:"column:number;" json:"number"`                               // 开的数字
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

// // 自定义JSON序列化功能
// func (wp NinePeriod) MarshalJSON() ([]byte, error) {
// 	type Alias NinePeriod
// 	return json.Marshal(&struct {
// 		ID uint `json:"id"`
// 		*Alias
// 	}{
// 		ID:    wp.ID,
// 		Alias: (*Alias)(&wp),
// 	})
// }

// NineOrder 订单表

type NineOrder struct {
	// BaseModel
	BaseModel
	UID          uint    `gorm:"column:uid;index:idx_uid_bet_type"`                                     // 用户ID 注意索引的名称和排序字段
	BetType      uint8   `gorm:"column:bet_type;index:idx_uid_bet_type"  json:"betType"`                // 房间类型  同时作为复合索引的一部分
	PeriodID     string  `gorm:"column:period;size:35;" json:"periodID"`                                // 期数
	TicketNumber string  `gorm:"column:ticket_number;size:18" json:"ticketNumber"`                      // 彩票ID 选的数字
	Number       int8    `gorm:"column:number;default:0" json:"number"`                                 // 实际开奖数字
	Rate         uint8   `json:"rate"`                                                                  // 抽水比例
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

func (o *NineOrder) CalculateFee() { //抽水处理
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
type NineOrderReq struct {
	UID          uint   `json:"uid"  `                         // 平台用户ID
	PeriodID     string `json:"periodID"  binding:"required"`  // 期数ID，omitempty表示若字段为空，则不显示在JSON中
	TicketNumber string `json:"ticketNumber"  `                // 彩票数字，注意这里JSON标记要与struct中的字段名称匹配
	BetAmount    uint   `json:"betAmount"  binding:"required"` // 投注金额
	BetType      uint8  `json:"betType"  binding:"required"`   // 房间类型
}

func (o *NineOrderReq) CheckTicketInvalid() error {
	ss := strings.Split(o.TicketNumber, ",")
	z := regexp.MustCompile(`^[0-9]{1}$`)
	repeatTest := make(map[string]bool)

	for _, element := range ss {
		if !z.MatchString(element) { // 不是0-9的数字
			return errors.WithCode(errors.InvalidParam)
		}
		if _, ok := repeatTest[element]; ok { // 数字不能重复
			return errors.WithCode(errors.InvalidParam)
		} else {
			repeatTest[element] = true
		}
	}
	return nil
}

// 是否选了9个数
func (o *NineOrderReq) IsTicketNine() bool {
	return len(strings.Split(o.TicketNumber, ",")) == 9
}

type NineRoomResp struct {
	ID          uint             `json:"id"`          // ID
	Setting     *NineRoomSetting `json:"setting"`     // 设置
	PeriodID    string           `json:"periodID"`    // 期数ID
	StateSTime  int64            `json:"stateSTime"`  // 状态时间
	RoundSTime  int64            `json:"roundSTime"`  // 回合时间
	NowSTime    int64            `json:"nowSTime"`    // 当前时间
	PlayerCount int              `json:"playerCount"` // 玩家数目
	State       string           `json:"state"`       // 状态
	PeriodIndex uint             `json:"periodIndex"` // 期数索引
}

type NineOrderHistoryReq struct {
	Paginator
	UID     uint
	BetType uint8 `json:"betType" binding:"required"` // 房间类型，可以为空
}

type SimulateSettleOrdersReq struct {
	PeriodID string `json:"periodID"`                   // 期数ID
	BetType  uint8  `json:"betType" binding:"required"` // 房间类型，可以为空
}
