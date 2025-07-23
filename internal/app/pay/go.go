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
type GOPay struct {
	//
}

type GOResponse struct {
	Code    int32         `json:"code"` // Use int32 for integers.
	Message string        `json:"message"`
	Data    GOPaymentData `json:"data"`
}

func (tk *GOResponse) isSucc() bool {
	return tk.Code == 200
}

type GOPaymentData struct {
	MerchantNo  string `json:"merId"`
	OrderAmount int    `json:"amount"`
	OrderNo     string `json:"orderId"`
	TradeNo     int    `json:"id"`
	// Fee         float64 `json:"fee"`
	URL string `json:"payLink"`
}

type GOBalanceResponse struct {
	Code    int32         `json:"code"` // Use int32 for integers.
	Message string        `json:"message"`
	Data    GOBalanceData `json:"data"`
}

func (tk *GOBalanceResponse) isSucc() bool {
	return tk.Code == 200
}

type GOBalanceData struct {
	AvailableAmount float64 `json:"free"`
	MerchantNo      string  `json:"merId"`
	BalanceAmount   float64 `json:"total"`
}

func (p *GOPay) QueryBalance(params *PaymentParameters) (*BalanceResp, error) {
	data := map[string]string{
		"merId": params.MerNo,
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	if err != nil {
		return nil, err
	}

	var resp GOBalanceResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &BalanceResp{
			AvailableAmount: resp.Data.AvailableAmount / 100,
			BalanceAmount:   resp.Data.BalanceAmount / 100,
			FrozenAmount:    0,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *GOPay) RequestPaymentURL(params *PaymentParameters) (*PaymentResp, error) {
	data := map[string]string{
		"merId":     params.MerNo,
		"orderId":   params.MerOrderNo,
		"currency":  params.Currency, //
		"amount":    fmt.Sprintf("%0.f", params.OrderAmount*100),
		"notifyUrl": params.NotifyURL,
		"mobile":    params.Mobile,
		"type":      "1",            // 替换 $PPHONE
		"returnUrl": params.PageURL, // 替换为你的金额数据
		"userName":  params.Name,
		"email":     params.Email,
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	logger.ZInfo("GO RequestPaymentURL", zap.String("url", params.PlatformApiUrl), zap.Any("req", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp GOResponse
	err = structure.MapToStruct(result, &resp)
	logger.ZInfo("hello", zap.Any("result", result), zap.Error(err))
	logger.ZInfo("hello", zap.Any("haha", resp))
	if resp.isSucc() {
		return &PaymentResp{
			Url:     resp.Data.URL,
			TradeNo: fmt.Sprintf("%d", resp.Data.TradeNo),
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *GOPay) sign(data map[string]string, privateKey string) string {
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

	// fmt.Println("-------------------", queryString)

	// 进行MD5签名
	md5Hash := md5.New()
	md5Hash.Write([]byte(queryString))
	signature := md5Hash.Sum(nil)
	return strings.ToUpper(hex.EncodeToString(signature))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type GOBack struct {
	MerchantNo  string `json:"merId"` // 平台分配唯一的商户号
	TradeNo     string `json:"id"`
	OrderNo     string `json:"orderId"`     // 保证每笔订单唯一的商户单号
	OrderAmount string `json:"amount"`      // 订单金额【单位：元，float类型，保留2位小数】
	Fee         string `json:"fee"`         // 支付状态【0 未支付】【1 支付成功】
	Currency    string `json:"currency"`    // 秒级时间戳【10位】
	OperatorNum string `json:"operatorNum"` // 交易参考号
	Sign        string `json:"sign"`        // 签名，不存储在数据库中

	TradeStatus int
}

func (p *GOBack) sign(mkey string) string {
	// Construct key-value pairs, excluding the sign itself and any empty values.
	kv := map[string]string{
		"merId":       p.MerchantNo,
		"orderId":     p.OrderNo,
		"id":          p.TradeNo,
		"amount":      p.OrderAmount, // 格式化浮点数为字符串，保留两位小数
		"fee":         p.Fee,         // 格式化浮点数为字符串，保留两位小数
		"currency":    p.Currency,
		"operatorNum": p.OperatorNum,
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

	return strings.ToUpper(signature)
}

///////////////////////////////实现IPayBack 接口

func (p GOBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == strings.ToUpper(p.Sign)
}

func (p GOBack) IsTransactionSucc() bool {
	return p.TradeStatus == 1
}

func (p GOBack) GetTransferOrder() *TransferOrder {
	amount, _ := strconv.ParseFloat(p.OrderAmount, 64) // 将字符串转换为float64类型
	return &TransferOrder{
		MerOrderNo:  p.OrderNo,
		OrderNo:     p.TradeNo,
		OrderAmount: float64(amount / 100),
	}
}

func (p GOBack) GetSuccResp() string {
	return "SUCCESS"
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type GOWithdraw struct {
}

func (p *GOWithdraw) RequestWithdraw(params *WithdrawParameters) (*WithdrawResp, error) {
	data := map[string]string{
		"merId":     params.MerNo, // 替换 $PNAME
		"orderId":   params.MerOrderNo,
		"amount":    fmt.Sprintf("%.0f", params.OrderAmount*100),
		"notifyUrl": params.NotifyURL, //
		"type":      "1",              //
		"bankCode":  params.IFSC,

		"userName": params.Name,
		"account":  params.AccountNumber, // 替换回调URL
		"email":    params.Email,
		"mobile":   params.Mobile,
		"currency": params.Currency,
	}

	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	logger.ZInfo("GO RequestWithdraw", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp GOResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &WithdrawResp{
			TradeNo: fmt.Sprintf("%d", resp.Data.TradeNo),
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *GOWithdraw) sign(data map[string]string, privateKey string) string {
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

type WithdrawGOBack struct {
	MerchantNo  string `json:"merId"`    // 假设商户号最长20个字符
	OrderNo     string `json:"orderId"`  // 保证每笔订单唯一
	TradeNo     string `json:"id"`       // 平台订单号分配唯一
	OrderAmount string `json:"amount"`   // 订单金额【单位：元，float类型, 保留2位小数】
	Fee         string `json:"fee"`      // 实际支付金额【单位：元，float类型, 保留2位小数】
	Currency    string `json:"currency"` // 商户自定义参数
	Sign        string `json:"sign"`     // 假设签名长度不会fa93dbd2498a5d266ddf70092def4f2a
	AppKey      string `gorm:"-"`        // AppKey不是数据库字段，gorm:"-"表示忽略该字段
	TradeStatus int
}

func (p *WithdrawGOBack) sign(mkey string) string {
	params := map[string]string{
		"merId":    p.MerchantNo,
		"orderId":  p.OrderNo,
		"id":       p.TradeNo,
		"amount":   p.OrderAmount,
		"fee":      p.Fee,
		"currency": p.Currency,
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

	return strings.ToUpper(sign)
}

func (p WithdrawGOBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == strings.ToUpper(p.Sign)
}

func (p WithdrawGOBack) IsTransactionSucc() bool {
	return p.TradeStatus == 1
}

func (p WithdrawGOBack) GetTransferOrder() *TransferOrder {
	amount, _ := strconv.ParseFloat(p.OrderAmount, 64) // 将字符串转换为float64类型
	return &TransferOrder{
		MerOrderNo:  p.OrderNo,
		OrderNo:     p.TradeNo,
		OrderAmount: float64(amount / 100),
	}
}

func (p WithdrawGOBack) GetSuccResp() string {
	return "SUCCESS"
}
