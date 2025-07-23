package entities

import (
	"rk-api/pkg/math"

	"github.com/shopspring/decimal"
)

// QuizFetchRule 竞猜数据拉取规则
type QuizFetchRule struct {
	BaseModel    `json:"-"`
	Limit        uint   `gorm:"column:limit;default:0"`                  // 拉取数量
	StartDateMin string `gorm:"column:start_date_min;default:0;size:32"` // 开始时间最小值
	StartDateMax string `gorm:"column:start_date_max;default:0;size:32"` // 开始时间最大值
	EndDateMin   string `gorm:"column:end_date_min;default:0;size:32"`   // 结束时间最小值
	EndDateMax   string `gorm:"column:end_date_max;default:0;size:32"`   // 结束时间最大值
	Tags         string `gorm:"column:tags;default:0;size:32"`           // 标签
	VolumeMin    uint   `gorm:"column:volume_min;default:0"`             // 成交量最小值
	VolumeMax    uint   `gorm:"column:volume_max;default:0"`             // 成交量最大值
	IsFetch      uint8  `gorm:"column:is_fetch;default:0"`               // 是否拉取
}

func (t *QuizFetchRule) TableName() string {
	return "quiz_fetch_rule"
}

// QuizEvent 竞猜事件
type QuizEvent struct {
	BaseModel `json:"-"`
	EventID   uint    `gorm:"column:event_id;default:0"`  // 事件ID
	Slug      string  `gorm:"column:slug;size:512"`       // 事件标识
	Title     string  `gorm:"column:title;size:512"`      // 事件标题
	Icon      string  `gorm:"column:icon;size:512"`       // 事件图标
	Volume    float64 `gorm:"column:volume;default:0"`    // 事件成交量
	CloseAt   uint    `gorm:"column:close_at;default:0"`  // 事件结束时间
	IsClosed  uint8   `gorm:"column:is_closed;default:0"` // 事件是否关闭
	Winner    uint    `gorm:"column:winner;default:0"`    // 赢家市场ID
	Status    uint8   `gorm:"column:status;default:0"`    // 事件展示状态 0 不拉 1 拉取
	Priority  uint    `gorm:"column:priority;default:0"`  // 事件优先级 (较高数字表示较高优先级)
	IsFetch   uint8   `gorm:"column:is_fetch;default:0"`  // 是否拉取
	IsSettle  uint8   `gorm:"column:is_settle;default:0"` // 是否结算
}

func (t *QuizEvent) TableName() string {
	return "quiz_event"
}

// QuizMarket 竞猜事件市场
type QuizMarket struct {
	BaseModel      `json:"-"`
	EventID        uint    `gorm:"column:event_id;default:0"`                     // 事件ID
	MarketID       uint    `gorm:"column:market_id;default:0"`                    // 市场ID
	GroupItemTitle string  `gorm:"column:group_item_title;default:0;size:128"`    // 市场名称标题
	QuestionID     string  `gorm:"column:question_id;default:0;size:512"`         // 市场问题id
	ConditionID    string  `gorm:"column:condition_id;default:0;size:512"`        // 市场条件id
	Ratio          uint    `gorm:"column:ratio;default:0"`                        // 市场比例 两位百分比
	YesToken       string  `gorm:"column:yes_token;default:0;size:512"`           // Yes市场赔率token
	YesPrice       float64 `gorm:"column:yes_price;default:0;type:decimal(10,3)"` // Yes市场赔率
	NoToken        string  `gorm:"column:no_token;default:0;size:512"`            // No市场赔率token
	NoPrice        float64 `gorm:"column:no_price;default:0;type:decimal(10,3)"`  // No市场赔率
	IsYesWinner    uint8   `gorm:"column:is_yes_winner;default:0"`                // 结算是否Yes赢
}

func (t *QuizMarket) TableName() string {
	return "quiz_market"
}

// QuizInfoRsp 竞猜信息
type QuizInfoRsp struct {
	EventID  uint            // 事件ID
	Title    string          // 事件问题
	Icon     string          // 事件图标
	Volume   float64         // 事件成交量
	CloseAt  uint            // 事件结束时间
	IsClosed bool            // 事件是否关闭
	Markets  []*QuizInfoItem // 竞猜信息单个市场选项
}

// QuizInfoItem 竞猜信息单个市场选项
type QuizInfoItem struct {
	MarketID       uint    // 市场ID
	GroupItemTitle string  // 市场名称标题
	Ratio          uint    // 市场比例 两位百分比
	YesPrice       float64 // Yes市场赔率
	NoPrice        float64 // No市场赔率
	IsYesWinner    bool    // 结算是否Yes赢
}

// QuizListReq 竞猜列表请求
type QuizListReq struct {
	Paginator
}

// QuizEventData 竞猜事件数据
type QuizEventData struct {
	ID      string            `json:"id"`      // 事件ID
	Slug    string            `json:"slug"`    // 事件标识
	Title   string            `json:"title"`   // 事件标题
	Icon    string            `json:"icon"`    // 事件图标
	Volume  float64           `json:"volume"`  // 事件成交量
	EndDate string            `json:"endDate"` // 事件结束时间
	Markets []*QuizMarketData `json:"markets"` // 事件市场选项
}

// QuizMarketData 竞猜事件市场数据
type QuizMarketData struct {
	ID             string `json:"id"`             // 市场ID
	GroupItemTitle string `json:"groupItemTitle"` // 市场名称标题
	QuestionID     string `json:"questionId"`     // 市场问题id
	ConditionID    string `json:"conditionId"`    // 市场条件id
	ClobTokenIds   string `json:"clobTokenIds"`   // 市场赔率token
	OutcomePrices  string `json:"outcomePrices"`  // 市场赔率
}

// QuizBuyRecord 竞猜购买记录
type QuizBuyRecord struct {
	BaseModel      `json:"-"`
	UID            uint    `json:"uid" gorm:"column:uid;default:0"`                                     // 用户ID
	EventID        uint    `json:"eventID" gorm:"column:event_id;default:0"`                            // 事件ID
	MarketID       uint    `json:"marketID" gorm:"column:market_id;default:0"`                          // 市场ID
	Title          string  `json:"title" gorm:"column:title;size:512"`                                  // 事件标题
	Icon           string  `json:"icon" gorm:"column:icon;size:512"`                                    // 事件图标
	GroupItemTitle string  `json:"groupItemTitle" gorm:"column:group_item_title;default:0;size:128"`    // 市场名称标题
	IsYes          uint8   `json:"isYes" gorm:"column:is_yes;default:0"`                                // 是否购买Yes
	PayMoney       float64 `json:"payMoney" gorm:"column:pay_money;default:0;type:decimal(10,3)"`       // 购买金额
	Price          float64 `json:"price" gorm:"column:price;default:0;type:decimal(10,3)"`              // 购买赔率
	Rate           uint8   `gorm:"column:rate;default:0" json:"-"`                                      // 抽水比例
	Delivery       float64 `gorm:"column:delivery;default:0;type:decimal(10,2)" json:"delivery"`        // 下注减抽水
	Fee            float64 `gorm:"column:fee;default:0;type:decimal(10,2)" json:"fee"`                  // 抽水
	IsSettle       uint8   `json:"isSettle" gorm:"column:is_settle;default:0"`                          // 是否已结算
	IsWin          uint8   `json:"isWin" gorm:"column:is_win;default:0"`                                // 是否赢
	SettleMoney    float64 `json:"settleMoney" gorm:"column:settle_money;default:0;type:decimal(10,3)"` // 结算金额
	StartTime      uint    `json:"startTime" gorm:"column:start_time;default:0"`                        // 购买时间
	SettleTime     uint    `json:"settleTime" gorm:"column:settle_time;default:0"`                      // 结算时间
	PromoterCode   int     `gorm:"column:pc;default:0" json:"-"`
}

func (t *QuizBuyRecord) TableName() string {
	return "quiz_buy_record"
}

func (o *QuizBuyRecord) CalculateFee() { //抽水处理
	decimalBetAmount := decimal.NewFromFloat(o.PayMoney)
	decimalRate := decimal.NewFromFloat(float64(o.Rate) / 1000)
	decimalFee := decimalBetAmount.Mul(decimalRate)
	decimalDelivery := decimalBetAmount.Sub(decimalFee)
	fee := math.MustParsePrecFloat64(decimalFee.InexactFloat64(), 3)
	delivery := math.MustParsePrecFloat64(decimalDelivery.InexactFloat64(), 3)
	o.Delivery = delivery
	o.Fee = fee
}

// QuizBuyReq 竞猜购买Req
type QuizBuyReq struct {
	UID      uint    `json:"-"`
	EventID  uint    `json:"event_id" binding:"required"`  // 事件ID
	MarketID uint    `json:"market_id" binding:"required"` // 市场ID
	IsYes    uint8   `json:"is_yes"`                       // 是否购买Yes
	PayMoney float64 `json:"pay_money" binding:"required"` // 购买金额
}

// QuizPriceData 竞猜价格数据
type QuizPriceData struct {
	Price string `json:"price"` // 市场赔率
}

// QuizBuyRecordReq 竞猜购买记录Req
type QuizBuyRecordReq struct {
	Paginator
	UID uint `json:"-"`
}

// QuizPricesHistoryReq 竞猜价格历史Req
type QuizPricesHistoryReq struct {
	EventID uint `json:"event_id" binding:"required"`
}

// QuizPricesHistoryRsp 竞猜价格历史Rsp
type QuizPricesHistoryRsp struct {
	List []*QuizPricesHistoryRspItem `json:"list"`
}

// QuizPricesHistoryRspItem 竞猜价格历史Rsp
type QuizPricesHistoryRspItem struct {
	EventID        uint                         `json:"event_id"`
	MarketID       uint                         `json:"market_id"`
	GroupItemTitle string                       `json:"group_item_title"`
	History        []*QuizPricesHistoryDataItem `json:"history"`
}

// QuizMarketPricesHistoryReq 竞猜市场价格历史Req
type QuizMarketPricesHistoryReq struct {
	EventID  uint `json:"event_id" binding:"required"`  // 事件ID
	MarketID uint `json:"market_id" binding:"required"` // 市场ID
}

type QuizPricesHistoryData struct {
	History []*QuizPricesHistoryDataItem `json:"history"`
}

type QuizPricesHistoryDataItem struct {
	Time  uint64  `json:"t"`
	Price float64 `json:"p"`
}
