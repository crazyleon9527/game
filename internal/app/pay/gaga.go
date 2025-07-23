package pay

import (
	"crypto/md5"
	"errors"
	"fmt"
	"rk-api/pkg/http"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"strings"

	"encoding/hex"
	"sort"

	"go.uber.org/zap"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type GaGaPay struct {
	//
}

const GAGA_APP_ID = "66a4bfba57b697e04f026c7d"

type GaGaResponse struct {
	Code    int32           `json:"code"` // Use int32 for integers.
	Message string          `json:"msg"`
	Data    GaGaPaymentData `json:"data"`
}

func (tk *GaGaResponse) isSucc() bool {
	return tk.Code == 0 && (tk.Data.OrderState < 3) //订单状态 0-订单生成 1-支付中 2-支付成功 3-支付失败 4-已撤销 5-已退款 6-订单关闭
}

type GaGaPaymentData struct {
	ErrCode     string `json:"errCode"`
	ErrMsg      string `json:"errMsg"`
	OrderNo     string `json:"mchOrderNo"`
	OrderState  int    `json:"orderState"`
	TradeNo     string `json:"payOrderId"`
	PayDataType string `json:"payDataType"` //payUrl-跳转链接的方式 form-表单方式 codeUrl-二维码地址 codeImgUrl-二维码图片地址 none-空支付参数
	URL         string `json:"payData"`
}

type GaGaBalanceResponse struct {
	Code    int32           `json:"code"` // Use int32 for integers.
	Message string          `json:"msg"`
	Data    GaGaBalanceData `json:"data"`
}

func (tk *GaGaBalanceResponse) isSucc() bool {
	return tk.Code == 0
}

type GaGaBalanceData struct {
	MerchantNo    string  `json:"mchNo"`
	BalanceAmount float64 `json:"balance"`
	FrozenAmount  float64 `json:"agentBalance"`
	AppId         string  `json:"appId"`
	ErrCode       string  `json:"errCode"`
	ErrMsg        string  `json:"errMsg"`
}

func (p *GaGaPay) QueryBalance(params *PaymentParameters) (*BalanceResp, error) {
	data := map[string]string{
		"mchNo": params.MerNo,
		"appId": GAGA_APP_ID,
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	if err != nil {
		return nil, err
	}

	var resp GaGaBalanceResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &BalanceResp{
			AvailableAmount: resp.Data.BalanceAmount / 100,
			BalanceAmount:   resp.Data.BalanceAmount / 100,
			FrozenAmount:    0,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *GaGaPay) RequestPaymentURL(params *PaymentParameters) (*PaymentResp, error) {
	data := map[string]string{
		"mchNo":         params.MerNo,
		"appId":         GAGA_APP_ID,
		"mchOrderNo":    params.MerOrderNo,
		"amount":        fmt.Sprintf("%0.f", params.OrderAmount*100),
		"customerEmail": params.Email,
		"customerPhone": params.trimMobile(),
		"customerName":  params.Name,
		"notifyUrl":     params.NotifyURL,
		"extParam":      params.MerOrderNo,
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	logger.ZInfo("GaGa RequestPaymentURL", zap.String("url", params.PlatformApiUrl), zap.Any("req", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp GaGaResponse
	err = structure.MapToStruct(result, &resp)
	logger.ZInfo("hello", zap.Any("result", result), zap.Error(err))
	logger.ZInfo("hello", zap.Any("haha", resp))
	if resp.isSucc() {
		return &PaymentResp{
			Url:     resp.Data.URL,
			TradeNo: resp.Data.TradeNo,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *GaGaPay) sign(data map[string]string, privateKey string) string {
	// 对map的键进行排序
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建待签名的字符串
	var queryStringBuilder strings.Builder
	for _, k := range keys {
		if data[k] != "" { //非空参与签名
			queryStringBuilder.WriteString(k)
			queryStringBuilder.WriteString("=")
			queryStringBuilder.WriteString(data[k])
			queryStringBuilder.WriteString("&")
		}
	}

	// 去除最后一个"&"
	queryString := strings.TrimRight(queryStringBuilder.String(), "&")

	// 添加密钥
	queryString += "&key=" + privateKey

	// 进行MD5签名
	md5Hash := md5.New()
	md5Hash.Write([]byte(queryString))
	signature := md5Hash.Sum(nil)
	return strings.ToUpper(hex.EncodeToString(signature))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type GaGaBack struct {
	PayOrderId     string `form:"payOrderId" json:"payOrderId"`                             // 代收订单号
	MchNo          string `form:"mchNo" json:"mchNo"`                                       // 商户号
	AppId          string `form:"appId" json:"appId"`                                       // 应用编号
	WayCode        string `form:"wayCode,omitempty" json:"wayCode,omitempty"`               // 支付方式
	IfCode         string `form:"ifCode,omitempty" json:"ifCode,omitempty"`                 // 支付编码
	ClientIp       string `form:"clientIp,omitempty" json:"clientIp,omitempty"`             // 客户端IP
	MchOrderNo     string `form:"mchOrderNo" json:"mchOrderNo"`                             // 商户订单号
	ChannelOrderNo string `form:"channelOrderNo,omitempty" json:"channelOrderNo,omitempty"` // 渠道订单号
	Amount         int    `form:"amount" json:"amount"`                                     // 支付金额（单位分）
	Currency       string `form:"currency" json:"currency"`                                 // 货币代码，固定值 INR
	CustomerName   string `form:"customerName" json:"customerName"`                         // 客户姓名
	CustomerEmail  string `form:"customerEmail" json:"customerEmail"`                       // 客户邮箱
	CustomerPhone  string `form:"customerPhone" json:"customerPhone"`                       // 客户手机号
	State          int    `form:"state" json:"state"`                                       // 订单状态 2-成功 3-失败
	ErrCode        string `form:"errCode,omitempty" json:"errCode,omitempty"`               // 渠道错误码
	ErrMsg         string `form:"errMsg,omitempty" json:"errMsg,omitempty"`                 // 渠道错误描述
	CreatedAt      int64  `form:"createdAt" json:"createdAt"`                               // 创建时间，13 位时间戳
	ReqTime        int64  `form:"reqTime" json:"reqTime"`                                   // 通知请求时间，13 位时间戳
	ExtParam       string `form:"extParam,omitempty" json:"extParam,omitempty"`             // 扩展参数
	Sign           string `form:"sign" json:"sign"`                                         // 签名
}

func (p *GaGaBack) sign(mkey string) string {
	kv := map[string]string{
		"payOrderId":     p.PayOrderId,
		"mchNo":          p.MchNo,
		"appId":          p.AppId,
		"wayCode":        p.WayCode,
		"ifCode":         p.IfCode,
		"clientIp":       p.ClientIp,
		"mchOrderNo":     p.MchOrderNo,
		"channelOrderNo": p.ChannelOrderNo,
		"amount":         fmt.Sprintf("%d", p.Amount), // Convert integer to string
		"currency":       p.Currency,
		"customerName":   p.CustomerName,
		"customerEmail":  p.CustomerEmail,
		"customerPhone":  p.CustomerPhone,
		"state":          fmt.Sprintf("%d", p.State), // Convert integer to string
		"errCode":        p.ErrCode,
		"errMsg":         p.ErrMsg,
		"extParam":       p.ExtParam,
		"createdAt":      fmt.Sprintf("%d", p.CreatedAt), // Convert int64 to string
		"reqTime":        fmt.Sprintf("%d", p.ReqTime),   // Convert int64 to string
	}

	// Remove keys with empty values from the map.
	for key, value := range kv {
		if value == "" {
			delete(kv, key)
		}
	}

	// Sort the keys in ASCII ascending order.
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var signParts []string
	for _, k := range keys {
		signParts = append(signParts, k+"="+kv[k])
	}
	signStr := strings.Join(signParts, "&")

	signStr += "&key=" + mkey

	// logger.ZInfo("GaGaBack GaGa sign", zap.String("signStr", signStr))

	hash := md5.Sum([]byte(signStr))
	signature := hex.EncodeToString(hash[:])

	return strings.ToUpper(signature)
}

///////////////////////////////实现IPayBack 接口

func (p GaGaBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == strings.ToUpper(p.Sign)
}

func (p GaGaBack) IsTransactionSucc() bool {
	return p.State == 2
}

func (p GaGaBack) GetTransferOrder() *TransferOrder {
	return &TransferOrder{
		MerOrderNo:  p.MchOrderNo,
		OrderNo:     p.PayOrderId,
		OrderAmount: float64(p.Amount / 100),
	}
}

func (p GaGaBack) GetSuccResp() string {
	return "success"
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type GaGaWithdrawResponse struct {
	Code    int32            `json:"code"` // Use int32 for integers.
	Message string           `json:"msg"`
	Data    GaGaWithdrawData `json:"data"`
}

func (tk *GaGaWithdrawResponse) isSucc() bool {
	return tk.Code == 0 && (tk.Data.State < 3) // 0-订单生成 1-转账中 2-转账成功 3-转账失败 4-转账关闭
}

type GaGaWithdrawData struct {
	TransferId   string `json:"transferId"`   // 转账ID
	MchOrderNo   string `json:"mchOrderNo"`   // 商户订单号
	Amount       int    `json:"amount"`       // 支付金额
	MchFeeAmount int    `json:"mchFeeAmount"` // 商户手续费
	AmountTo     int    `json:"amountTo"`     // 实际到账金额
	AccountNo    string `json:"accountNo"`    // 账户号码
	AccountName  string `json:"accountName"`  // 账户名称
	State        int    `json:"state"`        // 状态
	ErrCode      string `json:"errCode"`      // 错误码
	ErrMsg       string `json:"errMsg"`       // 错误描述
}

type GaGaWithdraw struct {
}

func (p *GaGaWithdraw) RequestWithdraw(params *WithdrawParameters) (*WithdrawResp, error) {
	data := map[string]string{
		"mchNo":        params.MerNo,
		"appId":        GAGA_APP_ID,
		"mchOrderNo":   params.MerOrderNo,
		"amount":       fmt.Sprintf("%.0f", params.OrderAmount*100),
		"entryType":    "IMPS", //
		"accountNo":    params.AccountNumber,
		"accountCode":  params.IFSC,
		"bankName":     "",
		"accountName":  params.AccountName,
		"accountEmail": params.Email,
		"accountPhone": params.trimMobile(),
		"notifyUrl":    params.NotifyURL, //
		"transferDesc": params.MerOrderNo,
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	logger.ZInfo("GaGa RequestWithdraw", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp GaGaWithdrawResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &WithdrawResp{
			TradeNo: resp.Data.TransferId,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *GaGaWithdraw) sign(data map[string]string, privateKey string) string {
	// 对map的键进行排序
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建待签名的字符串
	var queryStringBuilder strings.Builder
	for _, k := range keys {
		if data[k] != "" {
			queryStringBuilder.WriteString(k)
			queryStringBuilder.WriteString("=")
			queryStringBuilder.WriteString(data[k])
			queryStringBuilder.WriteString("&")
		}
	}

	// 去除最后一个"&"
	queryString := strings.TrimRight(queryStringBuilder.String(), "&")

	// 添加密钥
	queryString += "&key=" + privateKey

	// 进行MD5签名
	hash := md5.Sum([]byte(queryString))
	signature := hex.EncodeToString(hash[:])

	return strings.ToUpper(signature)
}

type WithdrawGaGaBack struct {
	MchNo          string `form:"mchNo" json:"mchNo"`                   // 商户号
	AppId          string `form:"appId" json:"appId"`                   // 应用编号
	MchOrderNo     string `form:"mchOrderNo" json:"mchOrderNo"`         // 商户转账订单号
	TransferId     string `form:"transferId" json:"transferId"`         // 代付订单号
	Amount         int    `form:"amount" json:"amount"`                 // 到账金额，不含手续费，单位分
	MchFeeAmount   int    `form:"mchFeeAmount" json:"mchFeeAmount"`     // 手续费，单位分
	AmountTo       int    `form:"amountTo" json:"amountTo"`             // 交易总金额，含手续费，单位分
	Currency       string `form:"currency" json:"currency"`             // 货币代码，固定值 INR
	EntryType      string `form:"entryType" json:"entryType"`           // 入账方式，固定值 IMPS
	State          int    `form:"state" json:"state"`                   // 转账状态 2-成功 3-失败
	AccountNo      string `form:"accountNo" json:"accountNo"`           // 收款账号
	AccountName    string `form:"accountName" json:"accountName"`       // 收款人姓名
	AccountEmail   string `form:"accountEmail" json:"accountEmail"`     // 收款人邮箱
	AccountPhone   string `form:"accountPhone" json:"accountPhone"`     // 收款人手机号
	BankName       string `form:"bankName" json:"bankName"`             // 收款人开户行名称
	ChannelOrderNo string `form:"channelOrderNo" json:"channelOrderNo"` // 渠道订单号
	IfCode         string `form:"ifCode" json:"ifCode"`                 // 支付接口
	ErrCode        string `form:"errCode" json:"errCode"`               // 渠道错误码
	WayCode        string `form:"wayCode" json:"wayCode"`               // 支付编码
	ErrMsg         string `form:"errMsg" json:"errMsg"`                 // 渠道错误描述
	ExtraParam     string `form:"extraParam" json:"extraParam"`         // 商户扩展参数
	TransferDesc   string `form:"transferDesc" json:"transferDesc"`     // 转账描述
	CreatedAt      int64  `form:"createdAt" json:"createdAt"`           // 创建时间，13 位时间戳
	Sign           string `form:"sign" json:"sign"`                     // 签名
	ReqTime        int64  `form:"reqTime" json:"reqTime"`               // 请求时间，13 位时间戳
	AppKey         string `gorm:"-"`                                    // AppKey不是数据库字段，gorm:"-"表示忽略该字段
}

func (p *WithdrawGaGaBack) sign(mkey string) string {
	params := map[string]string{
		"mchNo":          p.MchNo,
		"appId":          p.AppId,
		"mchOrderNo":     p.MchOrderNo,
		"transferId":     p.TransferId,
		"amount":         fmt.Sprintf("%d", p.Amount),       // Convert integer to string
		"mchFeeAmount":   fmt.Sprintf("%d", p.MchFeeAmount), // Convert integer to string
		"amountTo":       fmt.Sprintf("%d", p.AmountTo),     // Convert integer to string
		"currency":       p.Currency,
		"entryType":      p.EntryType,
		"state":          fmt.Sprintf("%d", p.State), // Convert integer to string
		"accountNo":      p.AccountNo,
		"accountName":    p.AccountName,
		"accountEmail":   p.AccountEmail,
		"accountPhone":   p.AccountPhone,
		"bankName":       p.BankName,
		"channelOrderNo": p.ChannelOrderNo,
		"ifCode":         p.IfCode,
		"errCode":        p.ErrCode,
		"wayCode":        p.WayCode,
		"errMsg":         p.ErrMsg,
		"extraParam":     p.ExtraParam,
		"transferDesc":   p.TransferDesc,
		"createdAt":      fmt.Sprintf("%d", p.CreatedAt), // Convert int64 to string
		"reqTime":        fmt.Sprintf("%d", p.ReqTime),   // Convert int64 to string
	}

	// Exclude empty fields and "cancel_message" from signing
	var keys []string
	for k, v := range params {
		if v != "" {
			keys = append(keys, k)
		}
	}

	// Sort keys
	sort.Strings(keys)

	// Build query string
	var signStrings []string
	for _, k := range keys {
		signStrings = append(signStrings, k+"="+params[k])
	}
	signString := strings.Join(signStrings, "&")

	// Append the merchant key
	signString += "&key=" + mkey

	// logger.ZInfo("WithdrawGaGaBack sign", zap.String("signString", signString))

	// Create MD5 hash
	hasher := md5.New()
	hasher.Write([]byte(signString))
	sign := hex.EncodeToString(hasher.Sum(nil))

	return strings.ToUpper(sign)
}

func (p WithdrawGaGaBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == strings.ToUpper(p.Sign)
}

func (p WithdrawGaGaBack) IsTransactionSucc() bool {
	return p.State == 2
}

func (p WithdrawGaGaBack) GetTransferOrder() *TransferOrder {
	return &TransferOrder{
		MerOrderNo:  p.MchOrderNo,
		OrderNo:     p.TransferId,
		OrderAmount: float64(p.Amount / 100),
	}
}

func (p WithdrawGaGaBack) GetSuccResp() string {
	return "success"
}
