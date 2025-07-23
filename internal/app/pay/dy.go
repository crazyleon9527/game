package pay

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"rk-api/pkg/http"
	"rk-api/pkg/logger"
	"rk-api/pkg/structure"
	"strconv"
	"strings"
	"time"

	"sort"

	"go.uber.org/zap"
)

// const rsaPrivateKeyStr = "-----BEGIN PRIVATE KEY-----\nMIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAMOSQXwJuatLVwluCoOxaqfJA189QaG79DJ+1DcaLaMeN8+0HPowBjixaIiY71K423MAuKhnpFEZdfnwUzWvjJdN7rIkh/Z1vkU40cnE2VaXEgbyuxr5kT1QZz+KuBOJDijIOjVdt1Y1sZSDu8W7An77WoKMi99MwseBwWt75NU3AgMBAAECgYEAg5doio58qL5z3Pt3Ba+eBTGjHDU6cRnnsQZXwo/Rv1z4zr/xc4JW3VS58pd9CNsrhdEpbt712D/aewdy5b+uR34H89MYpXCRzSdLpzfTD+jmWXhYoUTBnPuR2WFvhVRAyiGQw8MoxNt2+caj8pwirE1jnqtv5gmraCY1/xDciPkCQQD1/y/Clbwssc1ea8YPjuUu7oaDw9PGJWlbkE91beWAc+erLLgPgXQfITfXLjE5a/+ZpW7XFhCBfX02YO0iSBxDAkEAy4YkdnugWraM09D5xzdYRoehfDsL7nQ05kHoefkdZBDQwoSyAFzuGR6HVY+cS+pcD7Ts+xSyyRpI9KNJ57CN/QJAZdV/9fN6dJ4eQCopUaN76JPBh6Z3cp1mIgt5eUlHKofQraHTiEe0xHZB4YgzxGua4gYD/nIZ3yENxocVY42qdQJBAJWDcSv9e/rIGsLMxYIdXWNK5k4OTqCZi/cPugpJANdvJv2PX/i2TE/1xnQLsUVv4LvFLUbymLj171yQzI1Bfb0CQC0xVcsgj5/Q19x1U/7ICDRoxNrUQAgEKLkwIcN2UmFYL6ce6doBpOnUpGrpIqY1XdVTaJEmMrJgW3tGKkgys3o=\n-----END PRIVATE KEY-----" // 你的RSA私钥字符串

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// https://seabird.world/api/order/pay/create
type DYPay struct {
	//
}

type DYResponse struct {
	Code    int           `json:"code"`
	Data    DYPaymentData `json:"data"`
	Message string        `json:"msg"`
}

func (DY *DYResponse) isSucc() bool {
	return DY.Code == 200
}

type DYPaymentData struct {
	MerNo      string `json:"merNo"`
	MerOrderNo string `json:"merOrderNo"`
	OrderID    string `json:"orderNo"`
	PaymentURL string `json:"orderData"`
}

type DYBalanceResponse struct {
	Code    int           `json:"code"`
	Data    DYBalanceData `json:"data"`
	Message string        `json:"msg"`
}

func (DY *DYBalanceResponse) isSucc() bool {
	return DY.Code == 200
}

type DYBalanceData struct {
	List []DYPlatformOrder `json:"list"` // 平台订单号列表
	Sign string            `json:"sign"` // 签名
}

type DYPlatformOrder struct {
	AccountNo     string `json:"accountNo"`     // 子账户
	Currency      string `json:"currency"`      // 币种
	Balance       string `json:"balance"`       // 余额
	FrozenBalance string `json:"frozenBalance"` // 冻结余额
}

func (o *DYPlatformOrder) UnmarshalJSON(data []byte) error {
	type Alias DYPlatformOrder
	aux := &struct {
		Balance       interface{} `json:"balance"`
		FrozenBalance interface{} `json:"frozenBalance"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	o.Balance = formatAsString(aux.Balance)
	o.FrozenBalance = formatAsString(aux.FrozenBalance)
	return nil
}

func formatAsString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.2f", val) // 使用两位小数
	case int:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// /api/bal
func (p *DYPay) QueryBalance(params *PaymentParameters) (*BalanceResp, error) {
	data := map[string]string{
		"merNo":     params.MerNo,
		"requestNo": fmt.Sprintf("%d", time.Now().UnixMilli()),
		"timestamp": fmt.Sprintf("%d", time.Now().UnixMilli()),
	}
	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	// fmt.Println(sign)
	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	logger.ZInfo("DY QueryBalance", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}
	var resp DYBalanceResponse
	structure.MapToStruct(result, &resp)

	if resp.isSucc() {
		balance, _ := strconv.ParseFloat(resp.Data.List[0].Balance, 64)             // 将字符串转换为float64类型
		frozenBalance, _ := strconv.ParseFloat(resp.Data.List[0].FrozenBalance, 64) // 将字符串转换为float64类型
		return &BalanceResp{
			AvailableAmount: balance,
			BalanceAmount:   balance,
			FrozenAmount:    frozenBalance,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *DYPay) RequestPaymentURL(params *PaymentParameters) (*PaymentResp, error) {
	params.Mobile = strings.TrimPrefix(params.Mobile, "+") //去掉前面的加号
	data := map[string]string{
		"merNo":       params.MerNo,
		"merOrderNo":  params.MerOrderNo,
		"name":        params.Name,
		"email":       params.Email,
		"phone":       params.Mobile,
		"orderAmount": fmt.Sprintf("%.2f", params.OrderAmount),
		"currency":    params.Currency,
		"busiCode":    "103001", // Assumed default value.
		"notifyUrl":   params.NotifyURL,
		"pageUrl":     params.PageURL,
		"timestamp":   fmt.Sprintf("%d", time.Now().UnixMilli()),
	}
	sign := p.sign(data, params.AppKey)
	data["sign"] = sign

	// fmt.Println(sign)

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	logger.ZInfo("DY RequestPaymentURL", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp DYResponse
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

func (p *DYPay) sign(data map[string]string, privateKey string) string {
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

	h := hmac.New(sha256.New, []byte(privateKey))
	h.Write([]byte(queryString))
	return fmt.Sprintf("%x", h.Sum(nil))

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type DYBack struct {
	OrderAmount string `json:"orderAmount"`
	PayAmount   string `json:"payAmount"`
	OrderNo     string `json:"orderNo" `
	MerNo       string `json:"merNo"`
	MerOrderNo  string `json:"merOrderNo"`
	PayTime     string `json:"payTime"`
	BusiCode    string `json:"busiCode"`
	TradeStatus int    `json:"status"`
	Sign        string `json:"sign"`
	AppKey      string `gorm:"-"`
}

func (DY *DYBack) sign(secretKey string) string {
	data := map[string]string{
		"orderAmount": DY.OrderAmount,
		"payAmount":   DY.PayAmount,
		"orderNo":     DY.OrderNo,
		"merNo":       DY.MerNo,
		"merOrderNo":  DY.MerOrderNo,
		"payTime":     DY.PayTime,
		"busiCode":    DY.BusiCode,
		"status":      fmt.Sprintf("%d", DY.TradeStatus),
	}

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

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(queryString))
	return fmt.Sprintf("%x", h.Sum(nil))
}

///////////////////////////////实现IPayBack 接口

func (p DYBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == p.Sign
}

func (p DYBack) IsTransactionSucc() bool {
	return p.TradeStatus == 5
}

func (p DYBack) GetTransferOrder() *TransferOrder {

	amount, _ := strconv.ParseFloat(p.OrderAmount, 64) // 将字符串转换为float64类型

	return &TransferOrder{
		MerOrderNo:  p.MerOrderNo,
		OrderNo:     p.OrderNo,
		OrderAmount: amount, //按分的。
	}
}

func (p DYBack) GetSuccResp() string {
	return "SUCCESS"
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type DYWithdraw struct {
}

type DYWithdrawResponse struct {
	Code    int            `json:"code"`
	Data    DYWithdrawData `json:"data"`
	Message string         `json:"msg"`
}

// 0待审核，9处理中，6提交失败，8代付失败
// 提款信息错误，直接提交就失败了，提交失败是没有回调的，所以您系统那边做一个状态判断哈，代付订单提交失败判定为失败
func (DY *DYWithdrawResponse) isSucc() bool {
	return DY.Code == 200 && DY.Data.Status != 6
}

type DYWithdrawData struct {
	MerOrderNo string `json:"merOrderNo"`
	OrderID    string `json:"orderNo"`
	Status     int    `json:"status"`
}

func (p *DYWithdraw) RequestWithdraw(params *WithdrawParameters) (*WithdrawResp, error) {
	params.Mobile = strings.TrimPrefix(params.Mobile, "+") //去掉前面的加号

	data := map[string]string{
		"bankCode":   "IMPS", //"UPI"
		"province":   params.IFSC,
		"accName":    params.AccountName,   // 姓名
		"accNo":      params.AccountNumber, // 银行账号或者UPI账号
		"busiCode":   "203001",
		"currency":   "INR",             // 货币类型
		"email":      params.Email,      // 邮箱
		"merNo":      params.MerNo,      // 商户编号id
		"merOrderNo": params.MerOrderNo, // 商户订单号，订单唯一

		"notifyUrl": params.NotifyURL, // 商户接受回调通知地址

		"orderAmount": fmt.Sprintf("%0.0f", params.OrderAmount),

		"phone":     params.Mobile, // 手机号
		"timestamp": fmt.Sprintf("%d", time.Now().UnixMilli()),
		// IFSC

	}

	sign := p.sign(data, params.AppKey, params.AppSecret)
	data["sign"] = sign

	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)

	logger.ZError("DY RequestWithdraw", zap.String("url", params.PlatformApiUrl), zap.Any("data", data), zap.Any("result", result), zap.Error(err))

	if err != nil {
		return nil, err
	}

	var resp DYWithdrawResponse
	structure.MapToStruct(result, &resp)
	if resp.isSucc() {
		return &WithdrawResp{
			TradeNo: resp.Data.OrderID,
		}, nil
	} else {
		return nil, errors.New(resp.Message)
	}
}

func (p *DYWithdraw) sign(data map[string]string, privateKey string, secretKey string) string {

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

	h := hmac.New(sha256.New, []byte(privateKey))
	h.Write([]byte(queryString))
	signHmacSHA256 := fmt.Sprintf("%x", h.Sum(nil))

	// 使用标准库中的RSA和base64进行加密

	rsaPrivateKeyStr := fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----", secretKey)

	rsaPrivateKey, err := LoadPrivateKeyFromString(rsaPrivateKeyStr)
	if err != nil {
		fmt.Println("Load private key error:", err)
	}

	rsaSign, err := RsaEncrypt(signHmacSHA256, rsaPrivateKey)
	if err != nil {
		fmt.Println("RSA Encrypt error:", err)
	}

	return rsaSign
}

// RsaEncrypt 使用RSA私钥对数据进行加密并返回base64编码的字符串
func RsaEncrypt(data string, privateKey *rsa.PrivateKey) (string, error) {

	hashedData := []byte(data)

	// 然后，使用私钥对哈希值进行签名
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.Hash(0), hashedData[:])
	if err != nil {
		return "", err
	}

	// 最后，将签名结果编码为Base64字符串
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	return encodedSignature, nil

}

// LoadPrivateKeyFromString 从字符串中加载RSA私钥
func LoadPrivateKeyFromString(privateKeyStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return nil, fmt.Errorf("private key error")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPrivateKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not RSA private key")
	}

	return rsaPrivateKey, nil
}

func RsaVerify(data string, encodedSignature string, publicKey *rsa.PublicKey) (bool, error) {
	// 首先将签名从Base64字符串解码成[]byte
	signature, err := base64.StdEncoding.DecodeString(encodedSignature)
	if err != nil {
		return false, err
	}

	// 创建哈希实例并计算数据的哈希值
	// 注意：您在上一个函数中使用的是直接数据的字节序列，而通常我们需要对数据计算哈希
	// 如果您在签名时也是这么处理的，那么在这里也使用相同的方式
	// 但如果您在签名时使用的是数据的哈希值，那么这里也应该使用哈希值
	// hashedData := sha256.Sum256([]byte(data))
	hashedData := []byte(data)

	// 使用公钥和提供的数据和签名来验证签名的有效性
	err = rsa.VerifyPKCS1v15(publicKey, crypto.Hash(0), hashedData[:], signature)
	if err != nil {
		// 如果验证失败，返回false和错误信息
		return false, err
	}

	// 如果没有错误，表示签名验证成功，返回true
	return true, nil
}

func LoadPublicKeyFromString(publicKeyStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil {
		return nil, fmt.Errorf("public key error")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPublicKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not RSA public key")
	}

	return rsaPublicKey, nil
}

type WithdrawDYBack struct {
	MerNo       string `json:"merNo"`       // 平台订单编号id
	MerOrderNo  string `json:"merOrderNo"`  // 商户编号id
	OrderNo     string `json:"orderNo"`     // 商户订单编号id
	OrderAmount string `json:"orderAmount"` // 金额（分）
	Sign        string `json:"sign"`        // 手续费（分）
	PayTime     string `json:"payTime"`     // 货币
	TradeStatus int    `json:"status"`      // 签名
	ResultCode  string `json:"resultCode"`  // 货币
	ResultMsg   string `json:"resultMsg"`   // 签名
	AppKey      string `gorm:"-"`           // AppKey不是数据库字段，gorm:"-"表示忽略该字段
}

func (DY *WithdrawDYBack) sign(mkey string) string {
	params := map[string]string{
		"merNo":       DY.MerNo,
		"merOrderNo":  DY.MerOrderNo,
		"orderNo":     DY.OrderNo,
		"orderAmount": DY.OrderAmount,
		"payTime":     DY.PayTime,
		"status":      fmt.Sprintf("%d", DY.TradeStatus),
		"resultCode":  DY.ResultCode,
		"resultMsg":   DY.ResultMsg,
	}

	// Exclude empty fields and "cancel_message" from signing
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Sort keys
	sort.Strings(keys)

	// 构建待签名的字符串
	var queryStringBuilder strings.Builder
	for _, k := range keys {
		if params[k] != "" {
			queryStringBuilder.WriteString(k)
			queryStringBuilder.WriteString("=")
			queryStringBuilder.WriteString(params[k])
			queryStringBuilder.WriteString("&")
		}
	}

	// 去除最后一个"&"
	queryString := strings.TrimRight(queryStringBuilder.String(), "&")

	h := hmac.New(sha256.New, []byte(mkey))
	h.Write([]byte(queryString))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (p WithdrawDYBack) VerifySignature(appKey string) bool {
	return p.sign(appKey) == p.Sign
}

func (p WithdrawDYBack) IsTransactionSucc() bool {
	return p.TradeStatus == 7
}

func (p WithdrawDYBack) GetTransferOrder() *TransferOrder {
	amount, _ := strconv.ParseFloat(p.OrderAmount, 64)
	return &TransferOrder{
		MerOrderNo:  p.MerOrderNo,
		OrderNo:     p.OrderNo,
		OrderAmount: amount,
	}
}

func (p WithdrawDYBack) GetSuccResp() string {
	return "SUCCESS"
}
