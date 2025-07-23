package pay

import "strings"

type IPayBack interface {
	// VerifySignature(appKey string) bool //校验是否成功
	IsTransactionSucc() bool          //是否交易成功
	GetTransferOrder() *TransferOrder // 交易单
	GetSuccResp() string              // 成功时给三方返回码 c.String(http.StatusOK, "SUCCESS")
}

type TransferOrder struct {
	MerOrderNo  string  `json:"mer_order_no"` //商户订单号
	OrderNo     string  `json:"order_no"`     //平台订单号
	OrderAmount float64 `json:"order_amount"` //交易金额
}

func (t *TransferOrder) GetMerOrderNo() string {
	return t.MerOrderNo
}
func (t *TransferOrder) GetOrderNo() string {
	return t.OrderNo
}
func (t *TransferOrder) GetOrderAmount() float64 {
	return t.OrderAmount
}

type IPay interface {
	RequestPaymentURL(param *PaymentParameters) (*PaymentResp, error)
	QueryBalance(param *PaymentParameters) (*BalanceResp, error)
}

type PaymentResp struct {
	Url     string `json:"url"`
	TradeNo string `json:"trade_no"`
}

type BalanceResp struct {
	AvailableAmount float64 `json:"available_amount"`
	FrozenAmount    float64 `json:"frozen_amount"`
	BalanceAmount   float64 `json:"balance_amount"`
}

// 基本的支付参数
type PaymentParameters struct {
	// BankCode    string `json:"bankCode"` //"PIX"
	// BusiCode    string `json:"busi_code"`    //支付类型 印度UPI
	UID            uint    `json:"uid"`    //平台玩家ID
	MerNo          string  `json:"mer_no"` //商户号
	Name           string  `json:"pname"`  //姓名
	Goods          string  `json:"goods"`  //商品名
	Email          string  `json:"email"`
	Mobile         string  `json:"phone"`
	OrderAmount    float64 `json:"order_amount"` //交易金额
	PageURL        string  `json:"pageUrl"`      //支付之后跳转地址（选填）
	NotifyURL      string  `json:"notifyUrl"`    //支付之后通知地址（选填）
	Currency       string  `json:"currency"`
	MerOrderNo     string  `json:"mer_order_no"`     //商户订单号
	PlatformApiUrl string  `json:"platform_api_url"` //支付平台访问地址
	AppKey         string  `json:"app_key"`          //签名的key
	AppSecret      string  `json:"app_secret"`       //

	Attach string `json:"attach"` //附加
}

// 兼容为空的
func (p *PaymentParameters) CheckCompatibility() *PaymentParameters {
	if p.Name == "" {
		p.Name = "DefaultName"
	}
	if p.Goods == "" {
		p.Goods = "DefaultGoods" //
	}
	if p.Email == "" {
		p.Email = "default@email.com" //
	}
	if p.Mobile == "" {
		p.Mobile = "1234567890" //
	}
	return p
}

func (p *PaymentParameters) trimMobile() string {
	return strings.TrimPrefix(p.Mobile, "+91") //去掉前面的加号
}

///////////////////////////////////////////////////////////////////////////////////////////////

type WithdrawResp struct {
	TradeNo string `json:"trade_no"`
}

type IWithdraw interface {
	RequestWithdraw(params *WithdrawParameters) (*WithdrawResp, error) //
}

type WithdrawParameters struct {
	MerOrderNo    string `json:"mer_order_no"` //商户订单号
	UID           uint   //用户ID
	IP            string // ip地址
	Name          string `json:"pname"` //账号 名
	IFSC          string //印度的金融号
	AccountNumber string //银行卡号
	AccountName   string //银行户名

	Email  string `json:"email"` //邮箱
	Mobile string `json:"phone"` //手机号

	OrderAmount float64 `json:"order_amount"` //交易金额（只支持整数）
	PageURL     string  `json:"pageUrl"`      //支付之后跳转地址（选填）
	NotifyURL   string  `json:"notifyUrl"`    //支付之后通知地址（选填）
	Currency    string  `json:"currency"`

	MerNo          string `json:"mer_no"`           //商户号
	PlatformApiUrl string `json:"platform_api_url"` //支付平台访问地址
	AppKey         string `json:"app_key"`          //签名的key
	AppSecret      string `json:"app_secret"`

	Attach string `json:"attach"`
}

// 兼容为空的
func (p *WithdrawParameters) CheckCompatibility() *WithdrawParameters {

	if p.Name == "" {
		p.Name = "DefaultName"
	}

	if p.Email == "" {
		p.Email = "default@email.com" //
	}
	if p.Mobile == "" {
		p.Mobile = "1234567890" //
	}

	p.PlatformApiUrl = strings.Replace(p.PlatformApiUrl, "\r", "", -1)
	p.PlatformApiUrl = strings.Replace(p.PlatformApiUrl, "\n", "", -1)

	return p
}

func (p *WithdrawParameters) trimMobile() string {
	return strings.TrimPrefix(p.Mobile, "+91") //去掉前面的加号
}
