package pay

import (
	"crypto"
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"rk-api/pkg/http"
	"rk-api/pkg/logger"
	"strings"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"sort"
)

const ST_KEY = "MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAJs6n2WLK8Zn43pVm//U/Si2rhWBrlYKd01dUQjeLiDKrJWd9wVijyvdGHPLUGMlUn3WqWiVnOb4CXvPXyZXuHtPljUaULbR0DFmz3ANayNEI4xdU5Q+B93Ue5YJ7Bi8mbC+ElzFyuuxZspPBixOB3PwoQN4F9FLlKH9bpzaklnxAgMBAAECgYBrOtvRcYoTzMA/SdQxrhgTf95RzPt5RFGVh9jqn1COJnOCB1UUyzjLvDegBdrKIoYRl6//Jxr0EnO6q023VvGAacQHiVMExbMATCnDFXyJMgiLvXKLnk3EvrZwAzkHS1ks/nNB/xh7kpoiMLw5zaCrcROCJX+xwC1TkIeu2RRoYQJBAMwll7gMhQ+7MnykPIoPa4op7dwoH+Mvi9mO9mpacqr53NkYcZI05iUh3uJLEr1elwZRsJHqOqeywtnkTAuTB78CQQDCqDRbYDUOccogiGUkofqvsj8np3HoQAhxkJJzI8CNqdLD/x5aHU/937237MGoZToBKSuU9062hoP/PcbreopPAkBjntCryshABfy8IDU+tgmncZCWR6pP5grbzszm115vmuCvvCLn0xKf+ihWy4XIjVkrhz+f5OpcnXpCdAq79zLnAkEAqKBYAtRcAfBXHkmp/MsJRIzQIwTuKzjVV7Pa+j19f/sep0VpQL1l31KkoiFKPhu63OiSZZC7smjjDgixOqrEBwJBAJQncgzLXfcslum7yCaq4bklrBNptGPaGWp1WuAw3tfo/+dWuqOOc87VzVPF0yX3YGIZPNkkJL/DBJhNsCpe1rs="

// chunkSplit模仿了PHP中的chunk_split函数
func chunkSplit(body string, limit int, end string) string {
	var charSlice []rune
	for _, r := range body {
		charSlice = append(charSlice, r)
	}

	var result string
	for len(charSlice) >= limit {
		result += string(charSlice[:limit]) + end
		charSlice = charSlice[limit:]
	}
	result += string(charSlice) + end

	return result
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type PolyPay struct {
	//
}

func (p *PolyPay) RequestPaymentURL(params *PaymentParameters) (*PaymentResp, error) {
	data := map[string]string{
		"bankCode":        "PIX",
		"mer_no":          params.MerNo, // 你需要替换 $APPID
		"pname":           params.Name,  // 替换 $PNAME
		"goods":           params.Goods,
		"pemail":          params.Email,
		"phone":           params.Mobile,                         // 替换 $PPHONE
		"order_amount":    fmt.Sprintf("%f", params.OrderAmount), // 替换为你的金额数据
		"timeout_express": "90m",
		"pageUrl":         params.PageURL,   // 替换你的返回URL
		"notifyUrl":       params.NotifyURL, // 替换回调URL
		"ccy_no":          "INR",
		"busi_code":       "100303",
		"mer_order_no":    params.MerOrderNo, // 替换 $oid
	}

	params.AppKey = ST_KEY //for test

	if _, err := p.sign(data, params.AppKey); err != nil {
		return nil, err
	}
	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)
	if err != nil {
		return nil, err
	}
	if result["status"] == "SUCCESS" {

		var paymentUrl PaymentResp
		paymentUrl.Url = result["order_data"].(string)
		return &paymentUrl, nil

	} else {
		return nil, errors.New(result["err_msg"].(string))
	}
}

func (p *PolyPay) sign(data map[string]string, privateKey string) (map[string]string, error) {
	// sort keys
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// create string from sorted data
	var str strings.Builder
	for _, k := range keys {
		str.WriteString(k)
		str.WriteString("=")
		str.WriteString(data[k])
		str.WriteString("&")
	}

	// trim trailing '&'
	strTrimmed := strings.TrimSuffix(str.String(), "&")

	pemData := "-----BEGIN PRIVATE KEY-----\n" + chunkSplit(privateKey, 64, "\n") + "-----END PRIVATE KEY-----\n"

	// 解码PEM格式的数据
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		log.Fatal("pem.Decode failed")
	}

	if block == nil {
		// 处理错误: 没有找到PEM数据块
		logger.Error("----------sign----2---------", privateKey)
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privateKeyRsa, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, err
	}

	hashed := sha256.Sum256([]byte(strTrimmed))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKeyRsa, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}

	// base64 encode
	signed := base64.StdEncoding.EncodeToString(signature)

	// replaces
	signedReplaced := strings.NewReplacer("+", "-", "/", "_", "=", "").Replace(signed)
	data["sign"] = signedReplaced

	return data, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PolyBack struct {
	BusiCode    string `json:"busi_code"`
	ErrCode     string `json:"err_code"`
	ErrMsg      string `json:"err_msg"`
	MerNo       string `json:"mer_no"`
	MerOrderNo  string `json:"mer_order_no"`
	OrderAmount string `json:"order_amount"`
	OrderNo     string `json:"order_no"`
	OrderTime   string `json:"order_time"`
	PayAmount   string `json:"pay_amount"`
	PayTime     string `json:"pay_time"`
	Status      string `json:"status"`
	Sign        string `json:"sign"`
	AppKey      string
}

func (p *PolyBack) md5sign(mkey string) string {
	var sign strings.Builder
	sign.WriteString("busi_code=" + p.BusiCode)
	if p.ErrCode != "" {
		sign.WriteString("&err_code=" + p.ErrCode)
	}
	if p.ErrMsg != "" {
		sign.WriteString("&err_msg=" + p.ErrMsg)
	}
	sign.WriteString("&mer_no=" + p.MerNo)
	sign.WriteString("&mer_order_no=" + p.MerOrderNo)
	sign.WriteString("&order_amount=" + p.OrderAmount)
	sign.WriteString("&order_no=" + p.OrderNo)
	sign.WriteString("&order_time=" + p.OrderTime)
	sign.WriteString("&pay_amount=" + p.PayAmount)
	sign.WriteString("&pay_time=" + p.PayTime)
	sign.WriteString("&status=" + p.Status)
	sign.WriteString("&key=" + mkey)

	return fmt.Sprintf("%x", md5.Sum([]byte(sign.String())))
}

///////////////////////////////实现IPayBack 接口

func (t *PolyBack) TableName() string {
	return "pay_poly_back"
}

func (p *PolyBack) VerifySignature() bool {
	return p.md5sign(p.AppKey) == p.Sign
}

func (p *PolyBack) IsTransactionSucc() bool {
	return p.Status == "SUCCESS"
}

func (p *PolyBack) GetMerOrderNo() string {
	return p.MerOrderNo
}

func (p *PolyBack) GetOrderNo() string {
	return p.OrderNo
}

func (p *PolyBack) GetOrderAmount() string {
	return p.OrderAmount
}

func (p *PolyBack) GetSuccResp() string {
	return "SUCCESS"
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PolyWithdraw struct {
}

func (p *PolyWithdraw) RequestWithdraw(params *WithdrawParameters) error {
	data := map[string]string{
		"summary":      "something goods",
		"bank_code":    "IMPS",       //支行代码  印度代付类型,UPI,IMPS,PAYTM,IMPS指银行卡代付；
		"mer_no":       params.MerNo, // 替换 $PNAME
		"mobile_no":    params.Mobile,
		"acc_name":     params.AccountName,
		"province":     params.IFSC,                           // 替换 $PPHONE
		"order_amount": fmt.Sprintf("%f", params.OrderAmount), // 替换为你的金额数据
		"acc_no":       params.AccountNumber,
		"notifyUrl":    params.NotifyURL, // 替换回调URL
		"identity_no":  params.IFSC,
		"ccy_no":       "INR",
		"mer_order_no": params.MerOrderNo, // 替换 $oid
	}

	if _, err := p.sign(data, params.AppKey); err != nil {
		return err
	}
	result, err := http.SendPost(http.GetHttpClient(), params.PlatformApiUrl, data, true)
	if err != nil {
		return err
	}
	if result["status"] == "SUCCESS" {
		return nil
	} else {
		return errors.New(result["err_msg"].(string))
	}
}

func (p *PolyWithdraw) sign(data map[string]string, privateKey string) (map[string]string, error) {
	// sort keys
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// create string from sorted data
	var str strings.Builder
	for _, k := range keys {
		str.WriteString(k)
		str.WriteString("=")
		str.WriteString(data[k])
		str.WriteString("&")
	}

	// trim trailing '&'
	strTrimmed := strings.TrimSuffix(str.String(), "&")

	// sign
	block, _ := pem.Decode([]byte(privateKey))
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privateKeyRsa, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, err
	}

	hashed := sha256.Sum256([]byte(strTrimmed))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKeyRsa, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}

	// base64 encode
	signed := base64.StdEncoding.EncodeToString(signature)

	// replaces
	signedReplaced := strings.NewReplacer("+", "-", "/", "_", "=", "").Replace(signed)
	data["sign"] = signedReplaced

	return data, nil
}

type WithdrawPolyBack struct {
	CcyNo       string `json:"ccy_no"`
	CreateTime  string `json:"create_time"`
	ErrCode     string `json:"err_code,omitempty"`
	ErrMsg      string `json:"err_msg,omitempty"`
	MerNo       string `json:"mer_no"`
	MerOrderNo  string `json:"mer_order_no"`
	OrderAmount string `json:"order_amount"`
	OrderNo     string `json:"order_no"`
	PayTime     string `json:"pay_time,omitempty"`
	Status      string `json:"status"`

	Sign   string `json:"sign"`
	AppKey string
}

func (p *WithdrawPolyBack) md5sign(mkey string) string {
	var sign strings.Builder
	sign.WriteString("ccy_no=" + p.CcyNo)
	sign.WriteString("&create_time=" + p.CreateTime)
	if p.ErrCode != "" {
		sign.WriteString("&err_code=" + p.ErrCode)
	}
	if p.ErrMsg != "" {
		sign.WriteString("&err_msg=" + p.ErrMsg)
	}
	sign.WriteString("&mer_no=" + p.MerNo)
	sign.WriteString("&mer_order_no=" + p.MerOrderNo)
	sign.WriteString("&order_amount=" + p.OrderAmount)
	sign.WriteString("&order_no=" + p.OrderNo)
	if p.PayTime != "" {
		sign.WriteString("&pay_time=" + p.PayTime)
	}
	sign.WriteString("&status=" + p.Status)
	sign.WriteString("&key=" + mkey)

	return fmt.Sprintf("%x", md5.Sum([]byte(sign.String())))
}

func (t *WithdrawPolyBack) TableName() string {
	return "withdraw_poly_back"
}

func (p *WithdrawPolyBack) VerifySignature() bool {
	return p.md5sign(p.AppKey) == p.Sign
}

func (p *WithdrawPolyBack) IsTransactionSucc() bool {
	return p.Status == "SUCCESS"
}

func (p *WithdrawPolyBack) GetMerOrderNo() string {
	return p.MerOrderNo
}

func (p *WithdrawPolyBack) GetOrderNo() string {
	return p.OrderNo
}

func (p *WithdrawPolyBack) GetOrderAmount() string {
	return p.OrderAmount
}

func (p *WithdrawPolyBack) GetSuccResp() string {
	return "SUCCESS"
}
