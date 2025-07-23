package pay

import (
	"crypto/md5"
	"errors"
	"fmt"
	"rk-api/pkg/http"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"strconv"
	"strings"

	"encoding/hex"
	"sort"

	"go.uber.org/zap"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// https://seabird.world/api/order/pay/create
type TKPay struct {
	//
}

type TKResponse struct {
	Code    int           `json:"code"`
	Data    TKPaymentData `json:"data"`
	Message string        `json:"message"`
}

func (tk *TKResponse) isSucc() bool {
	return tk.Code == 200
}

type TKPaymentData struct {
	Amount     int    `json:"amount"`
	Fee        int    `json:"fee"`
	ID         int    `json:"id"`
	MerchantID string `json:"merchant_id"`
	OrderID    string `json:"order_id"`
	PaymentURL string `json:"pay_url"`
}

type TKBalanceResponse struct {
	Code    int           `json:"code"`
	Data    TKBalanceData `json:"data"`
	Message string        `json:"message"`
}

func (tk *TKBalanceResponse) isSucc() bool {
	return tk.Code == 200
}

type TKBalanceData struct {
	TotalBalance     float64 `json:"total_balance"`
	MerchantID       string  `json:"merchant_id"`
	AvailableBalance float64 `json:"available_balance"`
}

// /api/bal
func (p *TKPay) QueryBalance(params *PaymentParameters) (*BalanceResp, error) {
	data := map[string]string{
		"merchant_id": params.MerNo,
	}
	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	// fmt.Println(sign)
	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	if err != nil {
		return nil, err
	}
	var resp TKBalanceResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &BalanceResp{
			AvailableAmount: resp.Data.AvailableBalance / 100,
			BalanceAmount:   resp.Data.TotalBalance / 100,
			FrozenAmount:    0,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *TKPay) RequestPaymentURL(params *PaymentParameters) (*PaymentResp, error) {
	data := map[string]string{
		"merchant_id": params.MerNo,
		"order_id":    params.MerOrderNo,
		"amount":      fmt.Sprintf("%.0f", params.OrderAmount*100),
		"pay_type":    "1", // Assumed default value.
		"notify_url":  params.NotifyURL,
		"return_url":  params.PageURL,
		"currency":    params.Currency,
		"name":        params.Name,
		"phone":       params.Mobile,
		"email":       params.Email,
	}
	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	// fmt.Println(sign)

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	logger.ZInfo("TK RequestPaymentURL", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp TKResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &PaymentResp{
			Url:     resp.Data.PaymentURL,
			TradeNo: resp.Data.OrderID,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *TKPay) sign(data map[string]string, privateKey string) string {
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
	md5Hash := md5.New()
	md5Hash.Write([]byte(queryString))
	signature := md5Hash.Sum(nil)
	return hex.EncodeToString(signature)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TKBack struct {
	TradeID     string `json:"id" gorm:"column:trade_id;type:varchar(20);not null"` // 平台订单编号id
	MerchantID  string `json:"merchant_id" gorm:"type:varchar(20);not null"`        // 商户编号id
	OrderID     string `json:"order_id" gorm:"type:varchar(30);not null"`           // 商户订单编号id
	Amount      string `json:"amount" gorm:"type:decimal(10,2);not null"`           // 金额（分）
	Fee         string `json:"fee" gorm:"type:decimal(10,2);"`                      // 手续费（分）
	Currency    string `json:"currency" gorm:"type:varchar(10);not null"`           // 货币
	OperatorNum string `json:"operator_num" gorm:"type:varchar(10)"`                // 运营商编号
	Sign        string `json:"sign" gorm:"-"`                                       // 签名

	AppKey      string `gorm:"-"`
	TradeStatus int
}

func (tk *TKBack) sign(secretKey string) string {
	kv := map[string]string{
		"id":           tk.TradeID,
		"merchant_id":  tk.MerchantID,
		"order_id":     tk.OrderID,
		"amount":       tk.Amount,
		"fee":          tk.Fee,
		"currency":     tk.Currency,
		"operator_num": tk.OperatorNum,
	}

	for key, value := range kv {
		if value == "" {
			delete(kv, key)
		}
	}

	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var signStr string
	for _, k := range keys {
		signStr += fmt.Sprintf("%s=%s&", k, kv[k])
	}

	signStr += "key=" + secretKey

	// logger.Error(signStr)

	md5Hasher := md5.New()
	md5Hasher.Write([]byte(signStr))
	signature := fmt.Sprintf("%x", md5Hasher.Sum(nil))
	return strings.ToUpper(signature)
}

///////////////////////////////实现IPayBack 接口

func (p TKBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == strings.ToUpper(p.Sign)
}

func (p TKBack) IsTransactionSucc() bool {
	return p.TradeStatus == 1
}

func (p TKBack) GetTransferOrder() *TransferOrder {

	amount, _ := strconv.ParseFloat(p.Amount, 64) // 将字符串转换为float64类型

	return &TransferOrder{
		MerOrderNo:  p.OrderID,
		OrderNo:     p.TradeID,
		OrderAmount: amount / 100, //按分的。
	}
}

func (p TKBack) GetSuccResp() string {
	return "SUCCESS"
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TKWithdraw struct {
}

type TKWithdrawResponse struct {
	Code    int            `json:"code"`
	Data    TKWithdrawData `json:"data"`
	Message string         `json:"message"`
}

func (tk *TKWithdrawResponse) isSucc() bool {
	return tk.Code == 200
}

type TKWithdrawData struct {
	Amount     int    `json:"amount"`
	Fee        int    `json:"fee"`
	ID         int    `json:"id"`
	MerchantID string `json:"merchant_id"`
	OrderID    string `json:"order_id"`
}

func (p *TKWithdraw) RequestWithdraw(params *WithdrawParameters) (*WithdrawResp, error) {
	data := map[string]string{
		"merchant_id":   params.MerNo,                                // 商户编号id
		"order_id":      params.MerOrderNo,                           // 商户订单号，订单唯一
		"amount":        fmt.Sprintf("%0.f", params.OrderAmount*100), // 订单金额，单位（分）
		"withdraw_type": "1",                                         // 银行转账：1   UPI转账：2
		"notify_url":    params.NotifyURL,                            // 商户接受回调通知地址
		"name":          params.AccountName,                          // 姓名
		"phone":         params.Mobile,                               // 手机号
		"email":         params.Email,                                // 邮箱
		"bank_code":     params.IFSC,                                 // IFSC
		"account":       params.AccountNumber,                        // 银行账号或者UPI账号
		"currency":      "INR",                                       // 货币类型
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	logger.ZError("TK RequestWithdraw", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp TKWithdrawResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &WithdrawResp{
			TradeNo: resp.Data.OrderID,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *TKWithdraw) sign(data map[string]string, privateKey string) string {

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

	// fmt.Println("-------------------", queryString)

	// 进行MD5签名
	md5Hash := md5.New()
	md5Hash.Write([]byte(queryString))
	signature := md5Hash.Sum(nil)
	return hex.EncodeToString(signature)
}

type WithdrawTKBack struct {
	TradeID     string `json:"id" gorm:"column:trade_id;type:varchar(20);"`  // 平台订单编号id
	MerchantID  string `json:"merchant_id" gorm:"type:varchar(20);not null"` // 商户编号id
	OrderID     string `json:"order_id" gorm:"type:varchar(30);not null"`    // 商户订单编号id
	Amount      string `json:"amount" gorm:"type:decimal(10,2);not null"`    // 金额（分）
	Fee         string `json:"fee" gorm:"type:decimal(10,2);"`               // 手续费（分）
	Currency    string `json:"currency" gorm:"type:varchar(10);not null"`    // 货币
	UTR         string `json:"utr" gorm:"type:varchar(10)"`                  // 运营商编号
	Sign        string `json:"sign" gorm:"-"`                                // 签名
	TradeStatus int
	AppKey      string `gorm:"-"` // AppKey不是数据库字段，gorm:"-"表示忽略该字段
}

func (tk *WithdrawTKBack) sign(mkey string) string {
	params := map[string]string{
		"id":          tk.TradeID,
		"merchant_id": tk.MerchantID,
		"order_id":    tk.OrderID,
		"amount":      tk.Amount,
		"fee":         tk.Fee,
		"currency":    tk.Currency,
		"utr":         tk.UTR,
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

	// Create MD5 hash

	md5Hasher := md5.New()
	md5Hasher.Write([]byte(signString))
	signature := fmt.Sprintf("%x", md5Hasher.Sum(nil))
	return strings.ToUpper(signature)
}

func (p WithdrawTKBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == strings.ToUpper(p.Sign)
}

func (p WithdrawTKBack) IsTransactionSucc() bool {
	return p.TradeStatus == 1
}

func (p WithdrawTKBack) GetTransferOrder() *TransferOrder {
	amount, _ := strconv.ParseFloat(p.Amount, 64)
	return &TransferOrder{
		MerOrderNo:  p.OrderID,
		OrderNo:     p.TradeID,
		OrderAmount: amount / 100,
	}
}

func (p WithdrawTKBack) GetSuccResp() string {
	return "success"
}
