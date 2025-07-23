package pay

import (
	"crypto/md5"
	"errors"
	"fmt"
	"rk-api/pkg/http"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"strings"
	"time"

	"encoding/hex"
	"sort"

	"go.uber.org/zap"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type KBPay struct {
	//
}

type KBResponse struct {
	Code    int32         `json:"code"` // Use int32 for integers.
	Message string        `json:"message"`
	Data    KBPaymentData `json:"data"`
}

func (tk *KBResponse) isSucc() bool {
	return tk.Code == 0
}

type KBPaymentData struct {
	MerchantNo  string  `json:"merchant_no"`  // Use string for string values.
	OrderAmount float64 `json:"order_amount"` // Use float64 for floating-point numbers.
	OrderNo     string  `json:"order_no"`     // Use string for string values.
	TradeNo     string  `json:"trade_no"`     // Use string for string values.
	URL         string  `json:"url"`          // Use string for URL strings.
}

type KBBalanceResponse struct {
	Code    int32         `json:"code"` // Use int32 for integers.
	Message string        `json:"message"`
	Data    KBBalanceData `json:"data"`
}

func (tk *KBBalanceResponse) isSucc() bool {
	return tk.Code == 0
}

type KBBalanceData struct {
	AvailableAmount float64 `json:"available_amount"`
	CurrencyName    string  `json:"currency_name"`
	CurrencyTag     string  `json:"currency_tag"`
	FrozenAmount    float64 `json:"frozen_amount"`
	MerchantNo      string  `json:"merchant_no"`
	BalanceAmount   float64 `json:"balance_amount"`
}

func (p *KBPay) QueryBalance(params *PaymentParameters) (*BalanceResp, error) {
	data := map[string]string{
		"merchant_no": params.MerNo,
		"timestamp":   fmt.Sprintf("%d", time.Now().Unix()),
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, false)

	if err != nil {
		return nil, err
	}

	var resp KBBalanceResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &BalanceResp{
			AvailableAmount: resp.Data.AvailableAmount,
			BalanceAmount:   resp.Data.BalanceAmount,
			FrozenAmount:    resp.Data.FrozenAmount,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *KBPay) RequestPaymentURL(params *PaymentParameters) (*PaymentResp, error) {
	data := map[string]string{
		"merchant_no":  params.MerNo,
		"order_no":     params.MerOrderNo,                       //
		"order_amount": fmt.Sprintf("%.2f", params.OrderAmount), // for test 100
		"notify_url":   params.NotifyURL,
		"timestamp":    fmt.Sprintf("%d", time.Now().Unix()),
		"payin_method": "1201",         // 替换 $PPHONE
		"return_url":   params.PageURL, // 替换为你的金额数据
		"order_attach": params.Attach,
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, false)

	logger.ZInfo("KB RequestPaymentURL", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp KBResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &PaymentResp{
			Url:     resp.Data.URL,
			TradeNo: resp.Data.TradeNo,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *KBPay) sign(data map[string]string, privateKey string) string {
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type KBBack struct {
	MerchantNo    string  `gorm:"column:merchant_no;type:varchar(50);not null" form:"merchant_no"`     // 平台分配唯一的商户号
	OrderNo       string  `gorm:"column:order_no;type:varchar(50);not null" form:"order_no"`           // 保证每笔订单唯一的商户单号
	TradeNo       string  `gorm:"column:trade_no;type:varchar(50);not null" form:"trade_no"`           // 平台分配唯一的平台单号
	OrderAmount   float64 `gorm:"column:order_amount;type:decimal(10,2);not null" form:"order_amount"` // 订单金额【单位：元，float类型，保留2位小数】
	TradeAmount   float64 `gorm:"column:trade_amount;type:decimal(10,2);not null" form:"trade_amount"` // 实际支付金额【单位：元，float类型，保留2位小数】
	TradeStatus   int     `gorm:"column:trade_status;type:int;not null" form:"trade_status"`           // 支付状态【0 未支付】【1 支付成功】
	Timestamp     int64   `gorm:"column:timestamp;type:bigint(10);not null" form:"timestamp"`          // 秒级时间戳【10位】
	TransferRefNo string  `gorm:"column:transfer_ref_no;type:varchar(50)" form:"transfer_ref_no"`      // 交易参考号
	OrderAttach   string  `gorm:"column:order_attach;type:varchar(500)" form:"order_attach"`           // 商户自定义参数
	Sign          string  `gorm:"-" form:"sign"`                                                       // 签名，不存储在数据库中
}

func (p *KBBack) sign(mkey string) string {
	// Construct key-value pairs, excluding the sign itself and any empty values.
	kv := map[string]string{
		"merchant_no":     p.MerchantNo,
		"order_no":        p.OrderNo,
		"trade_no":        p.TradeNo,
		"order_amount":    fmt.Sprintf("%.2f", p.OrderAmount), // 格式化浮点数为字符串，保留两位小数
		"trade_amount":    fmt.Sprintf("%.2f", p.TradeAmount), // 格式化浮点数为字符串，保留两位小数
		"trade_status":    fmt.Sprintf("%d", p.TradeStatus),   // 将整数状态转换为字符串
		"timestamp":       fmt.Sprintf("%d", p.Timestamp),     // 将整数时间戳转换为字符串
		"transfer_ref_no": p.TransferRefNo,
		"order_attach":    p.OrderAttach,
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

	hash := md5.Sum([]byte(signStr))
	signature := hex.EncodeToString(hash[:])

	return signature
}

///////////////////////////////实现IPayBack 接口

func (p KBBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == p.Sign
}

func (p KBBack) IsTransactionSucc() bool {
	return p.TradeStatus == 1
}

func (p KBBack) GetTransferOrder() *TransferOrder {
	return &TransferOrder{
		MerOrderNo:  p.OrderNo,
		OrderNo:     p.TradeNo,
		OrderAmount: p.OrderAmount,
	}
}

func (p KBBack) GetSuccResp() string {
	return "success"
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type KBWithdraw struct {
}

func (p *KBWithdraw) RequestWithdraw(params *WithdrawParameters) (*WithdrawResp, error) {
	data := map[string]string{

		"merchant_no":   params.MerNo, // 替换 $PNAME
		"order_no":      params.MerOrderNo,
		"order_amount":  fmt.Sprintf("%.2f", params.OrderAmount),
		"notify_url":    params.NotifyURL, //
		"payout_method": "10716",          //
		"account_name":  params.AccountName,
		"account_no":    params.AccountNumber, // 替换回调URL
		"timestamp":     fmt.Sprintf("%d", time.Now().Unix()),
		"order_attach":  params.Attach,
		"bank_code":     params.IFSC,

		"bank_sub":       "",
		"account_type":   "",
		"document_type":  "",
		"account_attach": "",
		"document_no":    "",
		"mobile_no":      params.Mobile,
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, false)

	logger.ZInfo("KB RequestWithdraw", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp KBResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &WithdrawResp{
			TradeNo: resp.Data.TradeNo,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *KBWithdraw) sign(data map[string]string, privateKey string) string {
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

	fmt.Println("-------------------", queryString)

	// 进行MD5签名
	md5Hash := md5.New()
	md5Hash.Write([]byte(queryString))
	signature := md5Hash.Sum(nil)
	return hex.EncodeToString(signature)
}

type WithdrawKBBack struct {
	MerchantNo    string  `form:"merchant_no" gorm:"type:varchar(20);not null;"`   // 假设商户号最长20个字符
	OrderNo       string  `form:"order_no" gorm:"type:varchar(30);not null;"`      // 保证每笔订单唯一
	TradeNo       string  `form:"trade_no" gorm:"type:varchar(32);not null;"`      // 平台订单号分配唯一
	OrderAmount   float64 `form:"order_amount" gorm:"type:decimal(10,2);not null"` // 订单金额【单位：元，float类型, 保留2位小数】
	TradeAmount   float64 `form:"trade_amount" gorm:"type:decimal(10,2);not null"` // 实际支付金额【单位：元，float类型, 保留2位小数】
	TradeStatus   int     `form:"trade_status" gorm:"type:int;not null"`           // 代付状态【0 下单成功】【1 处理中】【2 代付成功】【4 已取消】
	Timestamp     int     `form:"timestamp" gorm:"type:int;not null"`              // 时间戳为整数类型
	OrderAttach   string  `form:"order_attach" gorm:"type:varchar(32)"`            // 商户自定义参数
	CancelMessage string  `form:"cancel_message" gorm:"type:varchar(120)"`         // 取消原因
	Sign          string  `form:"sign" gorm:"type:varchar(64);not null"`           // 假设签名长度不会fa93dbd2498a5d266ddf70092def4f2a
	AppKey        string  `gorm:"-"`                                               // AppKey不是数据库字段，gorm:"-"表示忽略该字段
}

func (p *WithdrawKBBack) sign(mkey string) string {
	params := map[string]string{
		"merchant_no":    p.MerchantNo,
		"order_no":       p.OrderNo,
		"trade_no":       p.TradeNo,
		"order_amount":   fmt.Sprintf("%.2f", p.OrderAmount),
		"trade_amount":   fmt.Sprintf("%.2f", p.TradeAmount),
		"trade_status":   fmt.Sprintf("%d", p.TradeStatus),
		"timestamp":      fmt.Sprintf("%d", p.Timestamp),
		"order_attach":   p.OrderAttach,
		"cancel_message": p.CancelMessage,
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
	hasher := md5.New()
	hasher.Write([]byte(signString))
	sign := hex.EncodeToString(hasher.Sum(nil))

	return sign
}

func (p WithdrawKBBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == p.Sign
}

func (p WithdrawKBBack) IsTransactionSucc() bool {
	return p.TradeStatus == 2
}

func (p WithdrawKBBack) GetTransferOrder() *TransferOrder {
	return &TransferOrder{
		MerOrderNo:  p.OrderNo,
		OrderNo:     p.TradeNo,
		OrderAmount: p.OrderAmount,
	}
}

func (p WithdrawKBBack) GetSuccResp() string {
	return "success"
}
