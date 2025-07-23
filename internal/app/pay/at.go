package pay

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/utils"
	"rk-api/pkg/http"
	"rk-api/pkg/logger"
	"strconv"
	"strings"
	"time"

	"encoding/hex"
	"encoding/json"
	"sort"

	"go.uber.org/zap"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ATPay struct {
	//
}

type ATResponse struct {
	Code    string        `json:"code"` // Use int32 for integers.
	Message string        `json:"msg"`
	Data    ATPaymentData `json:"data"`
}

func (tk *ATResponse) isSucc() bool {
	return tk.Code == "00000"
}

type ATPaymentData struct {
	URL     string `json:"payUrl"` // Use string for URL strings.
	TradeNo string `json:"tradeNo"`
}

type ATBalanceResponse struct {
	Code    string        `json:"code"` // Use int32 for integers.
	Message string        `json:"msg"`
	Data    ATBalanceData `json:"data"`
}

func (tk *ATBalanceResponse) isSucc() bool {
	return tk.Code == "00000"
}

type ATBalanceData struct {
	AvailableAmount  float64 `json:"availableBanlance"` //withdrawBanlance  是否用这个
	Currency         string  `json:"currency"`
	FrozenAmount     float64 `json:"freezeBanlance"`
	WithdrawBanlance float64 `json:"withdrawBanlance"`
	BalanceAmount    float64 `json:"totalBanlance"`
}

func (p *ATPay) QueryBalance(params *PaymentParameters) (*BalanceResp, error) {

	client := http.GetHttpClient()

	header := map[string]string{
		"X-Qu-Signature-Version": "v1.0",
		"X-Qu-Signature-Method":  "HmacSHA256",
		"X-Qu-Nonce":             utils.NewNonce(),
		"X-Qu-Timestamp":         fmt.Sprintf("%d", time.Now().Unix()),
		"X-Qu-Access-Key":        params.AppKey,
		"X-Qu-Mid":               params.MerNo,
	}
	header["X-Qu-Signature"] = p.sign(header, params.AppSecret)

	result, err := client.R().SetHeaders(header).
		SetQueryParam("currency", "INR").
		Get(params.PlatformApiUrl)
	if err != nil {
		return nil, err
	}

	var resp ATBalanceResponse
	err = json.Unmarshal(result.Body(), &resp)
	if err != nil {
		return nil, err
	}
	if resp.isSucc() {
		return &BalanceResp{
			AvailableAmount: resp.Data.AvailableAmount,
			BalanceAmount:   resp.Data.BalanceAmount,
			FrozenAmount:    resp.Data.FrozenAmount,
		}, nil
	} else {
		return nil, errors.With(resp.Message)
	}
}

func (p *ATPay) RequestPaymentURL(params *PaymentParameters) (*PaymentResp, error) {

	client := http.GetHttpClient()

	header := map[string]string{
		"X-Qu-Signature-Version": "v1.0",
		"X-Qu-Signature-Method":  "HmacSHA256",
		"X-Qu-Nonce":             utils.NewNonce(),
		"X-Qu-Timestamp":         fmt.Sprintf("%d", time.Now().Unix()),
		"X-Qu-Access-Key":        params.AppKey,
		"X-Qu-Mid":               params.MerNo,
	}
	header["X-Qu-Signature"] = p.sign(header, params.AppSecret)
	header["Content-Type"] = "application/json"

	data := map[string]string{
		"totalAmount": fmt.Sprintf("%.2f", params.OrderAmount),        // 订单金额，保留两位小数
		"outTradeNo":  params.MerOrderNo,                              // 调用方自定义订单号
		"buyerId":     utils.EncodeMD5(fmt.Sprintf("%d", params.UID)), // 买家ID，这里假设您已通过某种方式将 userId 转换为 MD5
		"channelCode": "601",                                          // 充值渠道编号
		"notifyUrl":   params.NotifyURL,                               // 订单异步回调地址
		"payName":     params.Name,                                    // 充值用户真实姓名，如果有的话
	}

	// logger.ZError("RequestPaymentURL", zap.String("url", params.PlatformApiUrl), zap.Any("header", header), zap.Any("data", data))

	result, err := client.R().SetHeaders(header).
		SetBody(data).
		Post(params.PlatformApiUrl)

	logger.ZInfo("AT RequestPaymentURL", zap.String("url", params.PlatformApiUrl), zap.Any("header", header), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}
	var resp ATResponse

	err = json.Unmarshal(result.Body(), &resp)
	if err != nil {
		return nil, err
	}
	if resp.isSucc() {
		return &PaymentResp{
			Url: resp.Data.URL,
		}, nil
	} else {
		return nil, errors.With(resp.Message)
	}
}

func (p *ATPay) sign(params map[string]string, secret string) string {
	// 将参数名按照 ASCII 码从小到大排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接成字符串
	var stringToSign string
	for _, k := range keys {
		stringToSign += k + "=" + params[k] + "&"
	}

	stringToSign = strings.TrimRight(stringToSign, "&")

	// 使用HmacSHA256签名方法对待签名字符串进行签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	return strings.ToUpper(signature) // 转换为大写
}

// X-Qu-Access-Key=D6BC07JDRGCKTPJI6IZHUKF0X-Qu-Mid=563186X-Qu-Nonce=e11ba548ea84d0421586a609920a80aaX-Qu-Signature-Method=Hma        acSHA256X-Qu-Signature-Version=v1.0X-Qu-Timestamp=1708781778
//    X-Qu-Access-Key=8568DFGEO9EK853ROR0GD6LW&X-Qu-Mid=10668926&X-Qu-Nonce=pgb0cx6z03ashxe3waqaneau4hnqqssj&X-Qu-Signature-Method=HmacSHA256&X-Qu-Signature-Version=v1.0&X-Qu-Timestamp=1631959194

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ATBack struct {
	ServiceFee     string `json:"serviceFee"`
	TradeAmount    string `json:"tradeAmount"`
	TradeNo        string `json:"tradeNo"`
	TradeStatus    string `json:"tradeStatus"`
	OutTradeNo     string `json:"outTradeNo"`
	CurrencySymbol string `json:"currencySymbol"`
	EndTime        int64  `json:"endTime"`
}

func (p *ATBack) sign(params map[string]string, secret string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接成字符串
	var stringToSign string
	for _, k := range keys {
		stringToSign += k + "=" + params[k] + "&"
	}

	stringToSign = strings.TrimRight(stringToSign, "&")

	// 使用HmacSHA256签名方法对待签名字符串进行签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	return strings.ToUpper(signature) // 转换为大写
}

///////////////////////////////实现IPayBack 接口

func (p ATBack) VerifySignature(params map[string]string, secret, signatureReceived string) bool {
	return p.sign(params, secret) == strings.ToUpper(signatureReceived)
}

func (p ATBack) IsTransactionSucc() bool {
	return p.TradeStatus == "SUCCESS"
}

func (p ATBack) GetTransferOrder() *TransferOrder {
	amount, _ := strconv.ParseFloat(p.TradeAmount, 64)
	return &TransferOrder{
		MerOrderNo:  p.OutTradeNo,
		OrderNo:     p.TradeNo,
		OrderAmount: amount,
	}
}

func (p ATBack) GetSuccResp() string {
	return "OK"
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ATWithdraw struct {
}

func (p *ATWithdraw) RequestWithdraw(params *WithdrawParameters) (*WithdrawResp, error) {
	client := http.GetHttpClient()

	header := map[string]string{
		"X-Qu-Signature-Version": "v1.0",
		"X-Qu-Signature-Method":  "HmacSHA256",
		"X-Qu-Nonce":             utils.NewNonce(), // 使用 utils 包
		"X-Qu-Timestamp":         fmt.Sprintf("%d", time.Now().Unix()),
		"X-Qu-Access-Key":        params.AppKey,
		"X-Qu-Mid":               params.MerNo,
	}
	// 假设 p 是一个结构体，并且它有一个 sign 方法用来生成签名
	header["X-Qu-Signature"] = p.sign(header, params.AppSecret)
	header["Content-Type"] = "application/json"

	data := map[string]string{
		"currency":        "INR",                                   // CNY-人民币
		"tradeAmount":     fmt.Sprintf("%.2f", params.OrderAmount), // 提款金额，转换为字符串并保留两位小数
		"outTradeNo":      params.MerOrderNo,                       // 调用方自定义订单号
		"bankCardNo":      params.AccountNumber,                    // 收款银行卡号
		"bankName":        "bank",                                  // 银行名称
		"bankAccountName": params.AccountName,                      // 收款人姓名
		"bankBranchName":  "",                                      // 银行支行名称
		"bankNum":         params.IFSC,                             // IFSC编号
		"bankType":        "0",                                     // （0-银行卡帐号、2-UPI）
		"notifyUrl":       params.NotifyURL,
	}

	// logger.ZError("RequestWithdraw", zap.String("url", params.PlatformApiUrl), zap.Any("header", header), zap.Any("data", data))

	result, err := client.R().
		SetHeaders(header).
		SetBody(data).
		Post(params.PlatformApiUrl)
	if err != nil {
		return nil, err
	}

	logger.ZInfo("AT RequestWithdraw", zap.String("url", params.PlatformApiUrl), zap.Any("header", header), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	var resp ATResponse
	err = json.Unmarshal(result.Body(), &resp)
	if err != nil {
		return nil, err
	}
	if resp.isSucc() {
		return &WithdrawResp{
			TradeNo: resp.Data.TradeNo,
		}, nil
	} else {
		return nil, errors.With(resp.Message)
	}
}

func (p *ATWithdraw) sign(params map[string]string, secret string) string {
	// 将参数名按照 ASCII 码从小到大排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接成字符串
	var stringToSign string
	for _, k := range keys {
		stringToSign += k + "=" + params[k] + "&"
	}

	stringToSign = strings.TrimRight(stringToSign, "&")

	// 使用HmacSHA256签名方法对待签名字符串进行签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	return strings.ToUpper(signature) // 转换为大写
}

type WithdrawATBack struct {
	Currency        string `json:"currency"`         // 币种
	BankCode        string `json:"bankCode"`         // 银行代码
	BankAccountName string `json:"bankAccountName"`  // 银行账户名
	ServiceFee      string `json:"serviceFee"`       // 服务费用
	TradeAmount     string `json:"tradeAmount"`      // 交易金额
	TradeNo         string `json:"tradeNo"`          // 交易编号
	BankCardNo      string `json:"bankCardNo"`       // 银行卡号
	Utr             string `json:"utr"`              // UTR编号（独一无二的参考号）
	OutTradeNo      string `json:"outTradeNo"`       // 外部交易编号
	Remark          string `json:"remark"`           // 备注
	BankName        string `json:"bankName"`         // 银行名称
	EndTime         int64  `json:"endTime"`          // 结束时间（时间戳格式）
	ErrMsg          string `json:"errMsg,omitempty"` // 错误描述（非必填字段，omitempty 表示如果为空则忽略该字段）
	PayStatus       string `json:"payStatus"`        // 支付状态
}

func (p *WithdrawATBack) sign(params map[string]string, secret string) string {
	// 将参数名按照 ASCII 码从小到大排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接成字符串
	var stringToSign string
	for _, k := range keys {
		stringToSign += k + "=" + params[k] + "&"
	}

	stringToSign = strings.TrimRight(stringToSign, "&")

	// 使用HmacSHA256签名方法对待签名字符串进行签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	return strings.ToUpper(signature) // 转换为大写
}

func (p WithdrawATBack) VerifySignature(params map[string]string, secret, signatureReceived string) bool {
	return p.sign(params, secret) == strings.ToUpper(signatureReceived)
}

func (p WithdrawATBack) IsTransactionSucc() bool {
	return p.PayStatus == "SUCCESS"
}

func (p WithdrawATBack) GetTransferOrder() *TransferOrder {
	amount, _ := strconv.ParseFloat(p.TradeAmount, 64)
	return &TransferOrder{
		MerOrderNo:  p.OutTradeNo,
		OrderNo:     p.TradeNo,
		OrderAmount: amount,
	}
}

func (p WithdrawATBack) GetSuccResp() string {
	return "OK"
}
