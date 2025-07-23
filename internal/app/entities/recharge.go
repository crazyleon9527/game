package entities

import "encoding/json"

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////
// 充值商品表
type RechargeGood struct {
	BaseModel `json:"-"`
	Price     uint  `gorm:"column:price;default:0" json:"price"`   //
	Status    uint8 `gorm:"column:status;default:0" json:"status"` // 状态1是正常显示
}

// 充值订单表
type RechargeOrder struct {
	BaseModel    `json:"-"`
	OrderID      string  `gorm:"column:order_id;default:0;size:26;uniqueIndex" json:"orderID"`    // 订单号
	TradeID      string  `gorm:"column:pay_oid;default:0;size:32" json:"-"`                       // 商家订单号
	UID          uint    `gorm:"column:uid;default:0;index" json:"uid"`                           // 用户ID
	IP           string  `gorm:"column:ip;size:20" json:"-"`                                      // IP地址
	ProductID    string  `gorm:"column:pid;default:0;size:12" json:"-"`                           // 产品ID
	Price        float64 `gorm:"column:price;default:0;type:decimal(10,2)" json:"price"`          // 单价
	TotalAmount  float64 `gorm:"column:paymoney;default:0;type:decimal(10,2)" json:"totalAmount"` // 总金额
	Count        uint    `json:"count"`                                                           // 数量
	RechargeType uint8   `gorm:"column:recharge_type;default:0" json:"rechargeType"`              // 充值类型 0非首充 1首充60 2首充2000 3首充20000
	Channel      string  `gorm:"column:channel;size:20" json:"channel"`                           // 渠道
	Status       uint8   `gorm:"column:status;default:0" json:"status"`                           // 支付状态 0未支付 1支付成功 2取消 3支付失败 4谷歌已支付但未消耗
	StartTime    int64   `gorm:"column:start_time;default:0" json:"start_time"`                   //订单发起时间
	FinishTime   int64   `gorm:"column:finish_time;default:0" json:"finish_time"`                 //订单结束时间
	PromoterCode int     `gorm:"column:pc;default:0" json:"-"`
}

type RechargeActivity struct {
	BaseModel `json:"-"`
	OrderID   string `gorm:"column:order_id;default:0;size:26;uniqueIndex" json:"orderID"` // 订单号
	UID       uint   `gorm:"column:uid;default:0;index" json:"uid"`                        // 用户ID
	ActType   int8   `gorm:"column:act_type;default:0" json:"-"`                           //活动类型 0 无活动  1充值10000 送2000
	Status    uint8  `gorm:"column:status;default:0" json:"status"`                        // 支付状态 0未支付 1支付成功
}

// // 自定义JSON序列化功能
func (wp *RechargeOrder) MarshalJSON() ([]byte, error) {
	type Alias RechargeOrder
	return json.Marshal(&struct {
		CreatedAt int64 `json:"createdAt"`
		*Alias
	}{
		CreatedAt: wp.CreatedAt,
		Alias:     (*Alias)(wp),
	})
}

// 支付渠道状态表

type RechargeSetting struct {
	BaseModel     `json:"-"` // 在JSON中忽略这个嵌入的模型
	Name          string     `gorm:"column:name;size:20" json:"name"`
	RechargeState uint8      `gorm:"column:recharge_state;default:0" json:"-"`
	WithdrawState uint8      `gorm:"column:withdraw_state;default:0" json:"-"`

	AvailableAmount float64 `gorm:"column:available_amount;default:0;type:decimal(10,2)" json:"-"`
	FrozenAmount    float64 `gorm:"column:frozen_amount;default:0;type:decimal(10,2)" json:"-"`
	BalanceAmount   float64 `gorm:"column:balance_amount;default:0;type:decimal(10,2)" json:"-"`

	Sort uint8 `gorm:"column:sort" json:"-"`

	Status uint8 `gorm:"column:status;default:0" json:"status"`
}

func (t *RechargeSetting) TableName() string {
	return "recharge_setting"
}

// 支付配置信息表
type RechargeChannelSetting struct {
	BaseModel           `json:"-"`
	CID                 uint   `gorm:"column:pcid;default:0"`             //RechargeSetting 的ID
	Name                string `gorm:"column:name;size:20"`               //支付名称
	AppID               string `gorm:"column:app_id;size:255"`            //支付账号ID
	PayKey              string `gorm:"column:pay_key;size:255"`           //支付APP KEY
	PaySecret           string `gorm:"column:pay_secret;size:1000"`       //
	WithdrawKey         string `gorm:"column:withdraw_key;size:255"`      //
	PayCallBackUrl      string `gorm:"column:pay_callback_url;size:255"`  //支付回调地址
	PayReturnUrl        string `gorm:"column:pay_return_url;size:255"`    //支付返回地址
	WithdrawCallBackUrl string `gorm:"column:sett_callback_url;size:255"` //代付回调地址
	WithdrawReturnUrl   string `gorm:"column:sett_return_url;size:255"`   //代付返回地址
	Remark              string `gorm:"column:remark;size:255"`            //备注
	// Status              uint8  `gorm:"column:status;default:0"`           //状态 1正常 0删除'
	BalanceApiUrl  string `gorm:"column:balance_api_url;"`  //支付平台查询金额访问地址
	RechargeApiUrl string `gorm:"column:recharge_api_url;"` //支付平台支付访问地址
	WithdrawApiUrl string `gorm:"column:withdraw_api_url;"` //支付平台提现访问地址
}

type CompletedRecharge struct {
	ID           uint    `gorm:"primarykey" `
	UID          uint    `gorm:"column:uid"`
	TradeID      string  `gorm:"column:trade_id;default:0;size:32" json:"tradeID"`         // 支付平台订单号
	OrderID      string  `gorm:"column:order_id;default:0;size:30" json:"orderID"`         // 商家订单号
	Amount       float64 `gorm:"column:amount;default:0;type:decimal(10,2)" json:"amount"` // 总金额
	Channel      string  `gorm:"column:channel;size:20" json:"channel"`                    // 渠道
	CreateTime   int64   `gorm:"column:create_time"`
	TodayTime    int64   `gorm:"column:today_time"`
	PromoterCode int     `gorm:"column:pc;default:0" json:"-"`
	// remark     string  `gorm:"column:remark;size:50"`
}

type GetRechargeOrderListReq struct {
	Paginator
	UID uint
}

type RechargeConfig struct {
	Goods       []*RechargeGood    `json:"goods"`
	Channels    []*RechargeSetting `json:"channels"`
	MinRecharge float64            `json:"minRecharge"`
}

type GetRechargeUrlReq struct {
	UID     uint
	Name    string  `json:"name"`
	Cash    float64 `json:"cash"`
	ActType int8    `json:"actType"` //活动类型
}

type RechargeUrlInfo struct {
	Url string `json:"url"`
}
