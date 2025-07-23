package entities

import (
	"encoding/json"
	"rk-api/pkg/math"

	"github.com/shopspring/decimal"
)

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

type WithdrawCard struct {
	BaseModel
	UID           uint   `gorm:"column:uid;default:0;index" json:"uid"`
	IFSC          string `gorm:"column:ifsc;size:32;" json:"ifsc"`
	AccountNumber string `gorm:"column:account_number;size:20;unique;" json:"accountNumber"`
	Name          string `gorm:"column:name;size:20;" json:"name"`
	Active        uint8  `gorm:"column:active;default:0" json:"active"`          // 是否正在使用 0不是 1是
	Status        uint8  `gorm:"column:status;default:0" json:"status"`          // 状态 0删除 1正常
	Remark        string `gorm:"column:remark;size:64;" json:"remark,omitempty"` // 20220202 可选字段
	Email         string `gorm:"column:email;size:32;" json:"email,omitempty"`   // 20220202 可选字段
	Mobile        string `gorm:"column:mobile;size:18;" json:"mobile,omitempty"`
	City          string `gorm:"column:city;size:40;" json:"city,omitempty"`
	Address       string `gorm:"column:address;size:64;" json:"address,omitempty"`
	State         string `gorm:"column:state;size:64;" json:"state,omitempty"`
}

// 提现记录表
type HallWithdrawRecord struct {
	BaseModel     `json:"-"`
	OrderID       string  `gorm:"column:order_id;size:26;uniqueIndex" json:"orderID"`     // 订单号
	TradeID       string  `gorm:"column:trade_id;size:32;" json:"-"`                      // 交易号
	UID           uint    `gorm:"column:uid;default:0;index" json:"-"`                    // 用户ID
	Channel       string  `gorm:"column:channel;size:20" json:"channel"`                  // 渠道
	Cash          float64 `gorm:"column:cash;default:0;type:decimal(10,2)" json:"cash"`   // 提现金额
	Rate          uint    `gorm:"-" json:"-"`                                             // 汇率
	Fee           float64 `gorm:"column:fee;default:0;type:decimal(10,2)" json:"-"`       // 手续费
	RealCash      float64 `gorm:"column:real_cash;default:0;type:decimal(10,2)" json:"-"` // 实际到手现金
	IFSC          string  `gorm:"column:ifsc;size:32;" json:"ifsc"`                       // IFSC码 Indian Financial System Code 指定银行分行
	AccountNumber string  `gorm:"column:account_number;size:20;" json:"accountNumber"`    // 账户号码
	AccountName   string  `gorm:"column:account_name;size:20;" json:"accountName"`        // 账户名
	CheckUser     uint    `gorm:"column:check_user;default:0" json:"-"`                   // 审核人
	CheckTime     int64   `gorm:"column:check_time" json:"-"`                             // 审核时间
	GiveUser      uint    `gorm:"column:give_user;default:0" json:"-"`                    // 打款人
	GiveTime      int64   `gorm:"column:give_time;" json:"give_time"`                     // 打款时间
	Remark        string  `gorm:"column:remark;size:64;" json:"-"`                        // 备注
	Reason        string  `gorm:"column:reason;size:255;" json:"-"`                       // 驳回理由
	Status        uint8   `gorm:"column:status;default:0" json:"status"`                  // 状态 0待审核 1通过 2驳回 3打款成功 4打款失败

	StartTime    int64 `gorm:"column:start_time;default:0" json:"start_time"`   //订单发起时间
	FinishTime   int64 `gorm:"column:finish_time;default:0" json:"finish_time"` //订单结束时间
	PromoterCode int   `gorm:"column:pc;default:0" json:"-"`
}

// 自定义JSON序列化功能
func (wp *HallWithdrawRecord) MarshalJSON() ([]byte, error) {
	type Alias HallWithdrawRecord
	return json.Marshal(&struct {
		CreatedAt int64 `json:"createdAt"`
		*Alias
	}{
		CreatedAt: wp.CreatedAt,
		Alias:     (*Alias)(wp),
	})
}

func (o *HallWithdrawRecord) CalculateFee() { //抽水处理
	decimalCash := decimal.NewFromFloat(o.Cash)
	decimalRate := decimal.NewFromFloat(float64(o.Rate) / 100)
	decimalFee := decimalCash.Mul(decimalRate)
	decimalRealCash := decimalCash.Sub(decimalFee)
	fee := math.MustParsePrecFloat64(decimalFee.InexactFloat64(), 3) //保留2位
	realCash := math.MustParsePrecFloat64(decimalRealCash.InexactFloat64(), 3)
	o.Fee = fee
	o.RealCash = realCash
}

type CompletedWithdraw struct {
	ID           uint    `gorm:"primarykey" `
	UID          uint    `gorm:"column:uid"`
	TradeID      string  `gorm:"column:trade_id;default:0;size:32" json:"tradeID"`         // 支付平台订单号
	OrderID      string  `gorm:"column:order_id;default:0;size:30" json:"orderID"`         // 商家订单号
	Amount       float64 `gorm:"column:amount;default:0;type:decimal(10,2)" json:"amount"` // 总金额
	Channel      string  `gorm:"column:channel;size:20" json:"channel"`                    // 渠道
	CreateTime   int64   `gorm:"column:create_time"`
	TodayTime    int64   `gorm:"column:today_time"`
	PromoterCode int     `gorm:"column:pc;default:0" json:"-"`
}

// ////////////////////////////////////////////////////////////DB table ////////////////////////////////////////////////////////////////////////////////////////

type WithdrawDetail struct {
	Cards        []*WithdrawCard `json:"cards,omitempty"`
	MinWithdraw  float64         `json:"min_withdraw,omitempty"`
	MaxWithdraw  float64         `json:"max_withdraw,omitempty"`
	WithdrawCash float64         `json:"withdraw_cash,omitempty"`
	RechargeAll  float64         `json:"recharge_all,omitempty"`
}

type GetHallWithdrawRecordListReq struct {
	Paginator
	UID uint
}

type ApplyForWithdrawalReq struct {
	UID        uint
	Cash       float64
	VerifyCode string //暂时不需要
}

type AddWithdrawCardReq struct {
	UID           uint   `json:"uid"`
	IFSC          string `json:"ifsc"`
	AccountNumber string `json:"accountNumber"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Mobile        string `json:"mobile"`
	City          string `json:"city"`
	Address       string `json:"address"`
	State         string `json:"state"`
	VerCode       string `json:"verCode"`
}

type ReviewWithdrawalReq struct {
	ID      uint   `json:"id"`      // ID is the primary key identifier of the withdrawal request.
	SysUID  uint   `json:"sysUid"`  // SysUID refers to the system user ID associated with the request.
	Reason  string `json:"reason"`  // Reason is the explanation provided for the withdrawal.
	IP      string `json:"ip"`      // IP is the IP address from where the request was made.
	OptType uint8  `json:"optType"` // OptType indicates the type of operation performed.
}
