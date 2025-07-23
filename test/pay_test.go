package test

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
	"testing"
)

func sign(data map[string]string, privateKey string, secretKey string) string {

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

	fmt.Println("=============1=============", queryString)

	h := hmac.New(sha256.New, []byte(privateKey))
	h.Write([]byte(queryString))
	signHmacSHA256 := fmt.Sprintf("%x", h.Sum(nil))

	fmt.Println("============2==============", signHmacSHA256)

	// 使用标准库中的RSA和base64进行加密

	rsaPrivateKeyStr := fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----", secretKey)

	fmt.Println(rsaPrivateKeyStr)

	rsaPrivateKey, err := LoadPrivateKeyFromString(rsaPrivateKeyStr)
	if err != nil {
		fmt.Println("Load private key error:", err)
	}

	rsaSign, err := RsaEncrypt(signHmacSHA256, rsaPrivateKey)
	if err != nil {
		fmt.Println("RSA Encrypt error:", err)
	}

	fmt.Println("============3==============", rsaSign)

	rsapublicKeyStr := fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCaq2mUlNO+6q96GrQTPrhGsrvEuMnNeL+R7KAjtYnoN1aZ5v/AjHarCXGiP5+RH4d7vF0mC3fBs0puslgP7gcF8ev7AZyW8ACe6aIaqswDsZWg4Dv7EorxNjlOz5UPbazOMqhLH5XAMPhswgOArxHtl3I7BwBBsPDCEpxGjycpeQIDAQAB")

	fmt.Println(rsapublicKeyStr)

	publicKey, err := LoadPublicKeyFromString(rsapublicKeyStr)
	if err != nil {
		fmt.Println("Load public key error:", err)
		return ""
	}

	success, err := RsaVerify(signHmacSHA256, rsaSign, publicKey)
	if err != nil {
		fmt.Println("RSA Verify error:", err)
	}

	if success {
		fmt.Println("Signature verification successful")
	} else {
		fmt.Println("Signature verification failed")
	}

	return rsaSign
}

// RsaEncrypt 使用RSA私钥对数据进行加密并返回base64编码的字符串
func RsaEncrypt(data string, privateKey *rsa.PrivateKey) (string, error) {

	// 首先，创建哈希实例并计算数据的哈希值
	// hashedData := sha256.Sum256([]byte(data))

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

func TestSign(t *testing.T) {
	params := struct {
		IFSC          string
		AccountName   string
		AccountNumber string
		Email         string
		MerNo         string
		MerOrderNo    string
		NotifyURL     string
		OrderAmount   float64
		Mobile        string
	}{
		IFSC:          "IPOS0000001",
		AccountName:   "Hani",
		AccountNumber: "12341234",
		Email:         "default@email.com",
		MerNo:         "803090000155024",
		MerOrderNo:    "w80123981717791312315064",
		NotifyURL:     "https://api-dev.cheetahs.asia/api/withdraw/callback/dy",
		OrderAmount:   190.95,
		Mobile:        "3500000000",
	}

	// 使用map和fmt.Sprintf填充数据
	data := map[string]string{
		"bankCode":    "IMPS", // 或者 "UPI"
		"province":    params.IFSC,
		"accName":     params.AccountName,   // 姓名
		"accNo":       params.AccountNumber, // 银行账号或者UPI账号
		"busiCode":    "203001",
		"currency":    "INR",                                   // 货币类型
		"email":       params.Email,                            // 邮箱
		"merNo":       params.MerNo,                            // 商户编号id
		"merOrderNo":  params.MerOrderNo,                       // 商户订单号，订单唯一
		"notifyUrl":   params.NotifyURL,                        // 商户接受回调通知地址
		"orderAmount": fmt.Sprintf("%.2f", params.OrderAmount), // 订单金额
		"phone":       params.Mobile,                           // 手机号
		"timestamp":   "1717791357231",                         // 当前时间戳（毫秒）
	}

	privateKey := "BF690D9F10F60E316867DD9D1B57E42CDB20758446186629D9EECC62E9FD9B19"
	secretKey := `MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAJqraZSU077qr3oatBM+uEayu8S4yc14v5HsoCO1ieg3Vpnm/8CMdqsJcaI/n5Efh3u8XSYLd8GzSm6yWA/uBwXx6/sBnJbwAJ7pohqqzAOxlaDgO/sSivE2OU7PlQ9trM4yqEsflcAw+GzCA4CvEe2XcjsHAEGw8MISnEaPJyl5AgMBAAECgYAURmoVtxi2yy7rk7qNV0hyxBVHKW4SyERcjadEZxTH4xEwJY7bH86ihx9lRG/XZ0msV7niNdiiBK26KxjOJca3tZ0iF+TY2sZsfus7mOkxdqVEX1BRAasluf8DGdiJgQULVEMDDMHaIHjWRiufMD1F4JKfXiXYEKv1fY0u6Q5scQJBAMnRtTIXQNVivd+/MW8Teu0blnQIdyxr6IvINXZvDDiIJLpm/ywQW5QyY70HMr/TxBXQNp3CjGMoVStEedBKIrUCQQDEMUk09QdyZSoI+H/rsB6wbiaJBpliWNx4W6vAzIRO9lgzJFK48JynaarEJZg6Bgw/rHLSZOMWxiBGtVLr8FI1AkBr+bWeuhAm2jBJ4CnbiqmO596j78+KxaTh/FiWJ617JWO5EtfcxFeqvxbbkFlXhj33ibFe0DJ57p86ieU+ShutAkEArwTIqdVsr0BQH3CPrVGQDxQ0jEi2zGU5SKV+sp+/6DEavYTJxvHezfoVvKVNB3Ydty+/vrEBZG+am3lmX9QjgQJBAILOjzAy3PtRRyKLEEITndEXXQENFm4yWM4LsAiHCrSzH+FPRBjspIZYDECJEgVlRvUdvHtY9VnXgAcSOyQuRj8=`

	sign(data, privateKey, secretKey)

	// // 认定的测试数据，这里你需要替换成你的测试数据
	// testData := map[string]string{
	// 	"merchant_no":  "6022012",
	// 	"order_amount": "116.00",
	// 	"order_no":     "9815224257",
	// 	"trade_amount": "116.00",
	// 	"trade_status": "2",
	// 	"timestamp":    "1656252693",
	// 	"trade_no":     "R16562506264566045",
	// }

	// // 已知的正确签名结果
	// expectedSign := "ce09d1a58973c683dfcbe89763e8d7c1"

	// // 调用你的签名方法
	// sign := kbpay.Sign(testData, "7n4sl3BNvpkyrrSoQKTHXg16CdJdhVf6")

	// // 校验结果
	// if sign != expectedSign {
	// 	t.Errorf("得到的签名 (%v) 与预期的签名 (%v) 不匹配", sign, expectedSign)
	// }

	// unescape := `merchant_no=6022012&order_amount=116.00&order_no=9815224257&trade_amount=116.00&trade_status=2&timestamp=1656252693&trade_no=R16562506264566045`
	// sign_str := unescape + "&key=" + "7n4sl3BNvpkyrrSoQKTHXg16CdJdhVf6"
	// fmt.Println(Md5(sign_str))

	// paymentParameters := &pay.PaymentParameters{
	// 	MerNo: "202366100",
	// 	// Name:           user.Nickname,
	// 	// Email:          user.Email,
	// 	// Mobile:         user.Mobile,
	// 	OrderAmount: 100,
	// 	PageURL:     "https://www.google.com",
	// 	NotifyURL:   "https://www.google.com",
	// 	MerOrderNo:  "202366100160101",

	// 	Currency: "INR",

	// 	PlatformApiUrl: "",
	// 	AppKey:         "keyValue",
	// }

	// // {
	// // 	"amount": "10000",
	// // 	"currency": "INR",
	// // 	"merchant_id": "202366100",
	// // 	"notify_url": "https://www.google.com",
	// // 	"order_id": "202366100160101",
	// // 	"pay_type": "1",
	// // 	"return_url": "https://www.google.com",
	// // 	"sign": "待生成"
	// // }

	// pay := new(pay.TKPay)

	// paymentUrl, err := pay.RequestPaymentURL(paymentParameters)

	// fmt.Println(paymentUrl, err)

	// --header 'X-Qu-Signature-Version:v1.0' \
	// --header 'X-Qu-Nonce:rk72al4c1p4u01ts3uhmqwnnuxyy0zrx' \
	// --header 'X-Qu-Signature-Method:HmacSHA256' \
	// --header 'X-Qu-Timestamp:1631961772' \
	// --header 'X-Qu-Access-Key:K853ROR0GD6L' \
	// --header 'X-Qu-Mid:23341' \
	// --header 'X-Qu-Signature: F0C850ED9A700628D00F15E207B97D1F7D9359059085D3B32833AD7F20CCF4E4'

	// header := map[string]string{
	// 	"X-Qu-Signature-Version": "v1.0",
	// 	"X-Qu-Signature-Method":  "HmacSHA256",
	// 	"X-Qu-Nonce":             "g6rl7o2snl4rjbaei0gb0443ycwqbq0t",
	// 	"X-Qu-Timestamp":         "1708785047",
	// 	"X-Qu-Access-Key":        "D6BC07JDRGCKTPJI6IZHUKF0",
	// 	"X-Qu-Mid":               "563186",
	// }
	// header["X-Qu-Signature"] = sign(header, "DN2DOMLMVCM69NNHEYDPJVW19ZEU07QL")

	// fmt.Println(header["X-Qu-Signature"])

	// 	93701ECFDC61140A555DF3D3C5C8DA49B35F47B7491DF3D0136CF4D06E6EFBAB

	// 	"X-Qu-Signature":"DCCD74C5256838B67CADDAE3936118111DE7080ECE3001A086DED822565F12CC"}
	// "header":{"X-Qu-Access-Key":"D6BC07JDRGCKTPJI6IZHUKF0","X-Qu-Mid":"563186","X-Qu-Nonce":"g6rl7o2snl4rjbaei0gb0443ycwqbq0t","X-Qu-Signature-Method":"HmacSHA256","X-Qu-Timestamp":"1708785047"}

	// {"X-Qu-Access-Key":"D6BC07JDRGCKTPJI6IZHUKF0","X-Qu-Mid":"563186","X-Qu-Nonce":"g6rl7o2snl4rjbaei0gb0443ycwqbq0t","X-Qu-Signature-Method":"HmacSHA256","X-Qu-Timestamp":"1708785047"}

}

func TestQuery(t *testing.T) {
	// dy := pay.DYPay{}
	// params := pay.PaymentParameters{
	// 	PlatformApiUrl: "https://tasdf.dypap.com/payout/balanceQuery",
	// 	AppKey:         "keyValue",
	// 	MerNo:          "806090017823738",
	// 	MerOrderNo:     "202366100160101",
	// }
	// dy.QueryBalance(params)

	// params2 := pay.PaymentParameters{
	// 	PlatformApiUrl: "https://tasdf.dypap.com/payout/balanceQuery",
	// 	AppKey:         "keyValue",
	// 	MerNo:          "806090017823738",
	// }
	// structure.CopyIgnoreEmpty(&params2, &params)

	// tt := DYPlatformOrder{}

	// ss := map[string]interface{}{
	// 	"MerNo":   "806090017823738------",
	// 	"balance": 100,
	// }
	// fmt.Println(structure.MapToStruct(ss, &tt))

	// fmt.Println(tt.Balance)
}

// type DYPlatformOrder struct {
// 	AccountNo     string `json:"accountNo"`     // 子账户
// 	Currency      string `json:"currency"`      // 币种
// 	Balance       string `json:"balance"`       // 余额
// 	FrozenBalance string `json:"frozenBalance"` // 冻结余额
// }

// func sign(params map[string]string, secret string) string {
// 	// 将参数名按照 ASCII 码从小到大排序
// 	var keys []string
// 	for k := range params {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)

// 	// 拼接成字符串
// 	var stringToSign string
// 	for _, k := range keys {
// 		stringToSign += k + "=" + params[k] + "&"
// 	}

// 	stringToSign = strings.TrimRight(stringToSign, "&")

// 	// 使用HmacSHA256签名方法对待签名字符串进行签名
// 	h := hmac.New(sha256.New, []byte(secret))
// 	h.Write([]byte(stringToSign))
// 	signature := hex.EncodeToString(h.Sum(nil))

// 	return strings.ToUpper(signature) // 转换为大写
// }

// // amount=10000&currency=INR&merchant_id=202366100&notify_url=https://www.google.com&order_id=202366100160101&pay_type=1&return_url=https://www.google.com&key=keyValue
// // amount=10000&currency=INR&merchant_id=202366100&notify_url=https://www.google.com&order_id=202366100160101&pay_type=1&return_url=https://www.google.com&key=keyValue
// func Md5(s string) string {
// 	h := md5.New()
// 	h.Write([]byte(s))
// 	return hex.EncodeToString(h.Sum(nil))
// }
