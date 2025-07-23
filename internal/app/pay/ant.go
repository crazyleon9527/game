package pay

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/url"
	"rk-api/pkg/cjson"
	"rk-api/pkg/http"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"strconv"
	"strings"
	"time"

	"encoding/hex"
	"sort"

	"go.uber.org/zap"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type ANTPay struct {
	//
}

type ANTResponse struct {
	Code    int `json:"code"`
	Content struct {
		AccountName string `json:"accountName"`
		Upi         string `json:"upi"`
	} `json:"content"`
	Html    string `json:"html"`
	Msg     string `json:"msg"`
	OrderNo string `json:"orderNo"`
	PayUrl  string `json:"payUrl"`
	Qrcode  string `json:"qrcode"`
}

func (ANT *ANTResponse) isSucc() bool {
	return ANT.Code == 0
}

type ANTBalanceResponse struct {
	Status  bool           `json:"status"`
	Balance ANTBalanceData `json:"balance"`
	Message string         `json:"message"`
}

func (ANT *ANTBalanceResponse) isSucc() bool {
	return ANT.Status
}

type ANTBalanceData struct {
	MerchantCode        string `json:"merchant_code"`
	TotalOrderAmount    string `json:"total_order_amount"`
	CashingBalance      string `json:"cashing_balance"`
	TotalWithdrawAmount string `json:"total_withdraw_amount"`
	FrozenBalance       string `json:"frozen_balance"`
	Balance             string `json:"balance"`
	WithdrawBalance     string `json:"withdraw_balance"`
	UnsettledBalance    string `json:"unsettled_balance"`
}

func (p *ANTPay) QueryBalance(params *PaymentParameters) (*BalanceResp, error) {
	data := map[string]string{
		"merchant_code": params.MerNo,
	}

	sign := p.sign(data, params.AppKey)

	// 把data映射编码为JSON
	jsonData := cjson.StringifyIgnore(data)

	// 创建Payload结构体实例
	payload := ANTPayload{
		Signtype:  "MD5",
		Sign:      url.QueryEscape(sign),
		Transdata: url.QueryEscape(string(jsonData)),
	}

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, payload, true)
	logger.ZInfo("ANT QueryBalance", zap.String("url", params.PlatformApiUrl), zap.Any("req", data), zap.Any("result", result), zap.Error(err))
	if err != nil {
		return nil, err
	}

	var resp ANTBalanceResponse
	structure.MapToStruct(result, &resp)

	availableAmount, _ := strconv.ParseFloat(resp.Balance.WithdrawBalance, 64) // 将字符串转换为float64类型
	balanceAmount, _ := strconv.ParseFloat(resp.Balance.Balance, 64)           // 将字符串转换为float64类型
	frozenAmount, _ := strconv.ParseFloat(resp.Balance.FrozenBalance, 64)      // 将字符串转换为float64类型
	if resp.isSucc() {
		return &BalanceResp{
			AvailableAmount: availableAmount,
			BalanceAmount:   balanceAmount,
			FrozenAmount:    frozenAmount,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

type ANTPayload struct {
	Signtype  string `json:"signtype"`
	Sign      string `json:"sign"`
	Transdata string `json:"transdata"`
}

func (p *ANTPay) RequestPaymentURL(params *PaymentParameters) (*PaymentResp, error) {

	data := map[string]string{
		"merchant_code": params.MerNo,
		"order_no":      params.MerOrderNo,
		"order_amount":  fmt.Sprintf("%d", int(params.OrderAmount)),
		"order_time":    fmt.Sprintf("%d", time.Now().Unix()), //
		"product_name":  params.Goods + fmt.Sprintf("%d", int(params.OrderAmount)),
		"notify_url":    params.NotifyURL,
		"pay_type":      "india-upi-h5", // 替换 $PPHONE
		"return_url":    params.PageURL, // 替换为你的金额数据
		"payer_info":    params.Name,
	}

	sign := p.sign(data, params.AppKey)

	// 把data映射编码为JSON
	jsonData := cjson.StringifyIgnore(data)

	// 创建Payload结构体实例
	payload := ANTPayload{
		Signtype:  "MD5",
		Sign:      url.QueryEscape(sign),
		Transdata: url.QueryEscape(string(jsonData)),
	}

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, payload, true)

	logger.ZInfo("ANT RequestPaymentURL", zap.String("url", params.PlatformApiUrl), zap.Any("req", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp ANTResponse
	err = structure.MapToStruct(result, &resp)
	logger.ZInfo("hello", zap.Any("result", result), zap.Error(err))
	if resp.isSucc() {
		return &PaymentResp{
			Url:     resp.PayUrl,
			TradeNo: resp.OrderNo,
		}, nil
	} else {
		return nil, errors.New(resp.Msg)
	}
}

func (p *ANTPay) sign(data map[string]string, privateKey string) string {
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

type ANTBack struct {
	OrderNo     string `json:"order_no"`
	OrderAmount string `json:"order_amount"` // 注意这里是按字符串处理的，如果金额精度很关键需要做适当转换
	OrderTime   int64  `json:"order_time"`
	PayType     string `json:"pay_type"`
	ProductName string `json:"product_name"`
	ProductCode string `json:"product_code"`
	UserNo      string `json:"user_no"`
	Payment     string `json:"payment"`

	UTRCode string `json:"utr_code"`

	Sign string `json:"sign"` // 签名，不存储在数据库中
}

func (p *ANTBack) sign(mkey string) string {
	// Construct key-value pairs, excluding the sign itself and any empty values.
	kv := map[string]string{
		"order_no":     p.OrderNo,
		"order_amount": p.OrderAmount,
		"order_time":   fmt.Sprintf("%d", p.OrderTime),
		"pay_type":     p.PayType,     // 格式化浮点数为字符串，保留两位小数
		"product_name": p.ProductName, // 格式化浮点数为字符串，保留两位小数
		"product_code": p.ProductCode,
		"user_no":      p.UserNo,
		"payment":      p.Payment,
		"utr_code":     p.UTRCode,
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

func (p ANTBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == strings.ToUpper(p.Sign)
}

// 只有支付成功才会通知
func (p ANTBack) IsTransactionSucc() bool {
	return true
}

func (p ANTBack) GetTransferOrder() *TransferOrder {
	amount, _ := strconv.ParseFloat(p.OrderAmount, 64) // 将字符串转换为float64类型
	return &TransferOrder{
		MerOrderNo:  p.OrderNo,
		OrderNo:     p.OrderNo,
		OrderAmount: float64(amount),
	}
}

func (p ANTBack) GetSuccResp() string {
	return "SUCCESS"
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type ANTWithdrawResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

func (ANT *ANTWithdrawResponse) isSucc() bool {
	return ANT.Status
}

type ANTWithdraw struct {
}

func (p *ANTWithdraw) RequestWithdraw(params *WithdrawParameters) (*WithdrawResp, error) {
	data := map[string]string{
		"merchant_code": params.MerNo, // 替换 $PNAME
		"order_no":      params.MerOrderNo,
		"order_amount":  fmt.Sprintf("%d", int(params.OrderAmount)),

		"pay_type":    "india-bank-repay", //
		"bank_name":   params.IFSC,
		"bank_card":   params.AccountNumber, // 替换回调URL
		"bank_branch": params.IFSC,
		"user_name":   params.AccountName,
		"notifyUrl":   params.NotifyURL, //

	}

	sign := p.sign(data, params.AppKey)

	// 把data映射编码为JSON
	jsonData := cjson.StringifyIgnore(data)

	// 创建Payload结构体实例
	payload := ANTPayload{
		Signtype:  "MD5",
		Sign:      url.QueryEscape(sign),
		Transdata: url.QueryEscape(string(jsonData)),
	}

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, payload, true)

	logger.ZInfo("ANT RequestWithdraw", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp ANTWithdrawResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &WithdrawResp{
			TradeNo: fmt.Sprintf("%d", 0),
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *ANTWithdraw) sign(data map[string]string, privateKey string) string {
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

type WithdrawANTBack struct {
	OrderNo     string `json:"order_no"`
	OrderAmount string `json:"order_amount"`
	Message     string `json:"message"`
	RespCode    string `json:"resp_code"`
	UTRCode     string `json:"utr_code"`
	Sign        string `json:"sign"` // 假设签名长度不会fa93dbd2498a5d266ddf70092def4f2a
}

func (p *WithdrawANTBack) sign(mkey string) string {
	params := map[string]string{
		"order_no":     p.OrderNo,
		"order_amount": p.OrderAmount,
		"message":      p.Message,
		"resp_code":    p.RespCode,
		"utr_code":     p.UTRCode,
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

func (p WithdrawANTBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == strings.ToUpper(p.Sign)
}

func (p WithdrawANTBack) IsTransactionSucc() bool {
	return p.RespCode == "S" //S 代付成功，F 代付失败，P 代付中
}

func (p WithdrawANTBack) GetTransferOrder() *TransferOrder {
	amount, _ := strconv.ParseFloat(p.OrderAmount, 64) // 将字符串转换为float64类型
	return &TransferOrder{
		MerOrderNo:  p.OrderNo,
		OrderNo:     p.OrderNo,
		OrderAmount: float64(amount),
	}
}

func (p WithdrawANTBack) GetSuccResp() string {
	return "SUCCESS"
}
