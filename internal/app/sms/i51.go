package sms

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"rk-api/pkg/http"
	"time"
)

type I51Sms struct {
	appKey    string
	appSecret string
	apiUrl    string //"https://api.i51sms.com/outauth/verifCodeSend"
}

func NewI51Sms(appKey, appSecret, apiUrl string) IMessage {
	return &I51Sms{
		appKey:    appKey,
		appSecret: appSecret,
		apiUrl:    apiUrl,
	}
}

func (s *I51Sms) Send(phone, content string) (string, error) {
	// 获取当前时间戳并格式化为字符串
	timestamp := time.Now().Format("20060102150405")
	sign := generateMD5Sign(s.appKey, timestamp, s.appSecret)

	// 构建请求参数
	request := SmsRequest{
		Apikey:    s.appKey,
		Timestamp: timestamp,
		Sign:      sign,
		Mobile:    phone,
		Content:   content,
	}

	client := http.GetHttpClient()

	// 发起请求
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post(s.apiUrl)

	if err != nil {
		return "", err
	}

	// 解析响应
	var smsResponse SmsResponse
	if err := json.Unmarshal(resp.Body(), &smsResponse); err != nil {
		return "", err
	}

	if smsResponse.Code != "000" {
		// 打印响应结果
		fmt.Println("短信发送成功，响应信息如下:")
		fmt.Printf("状态: %s, 信息: %s, 错误码: %s, 数据: %+v\n",
			smsResponse.Status, smsResponse.Msg, smsResponse.Code, smsResponse.Data)

		return "", errors.New(smsResponse.Msg)
	}

	return fmt.Sprintf("%d", smsResponse.Data.Taskid), nil

}

// SmsRequest 定义了发送短信请求的结构体
type SmsRequest struct {
	Apikey     string `json:"apikey"`
	Timestamp  string `json:"timestamp"`
	Sign       string `json:"sign"`
	Mobile     string `json:"mobile"`
	Content    string `json:"content"`
	Senderid   string `json:"senderid,omitempty"`
	Templateid string `json:"templateid,omitempty"`
	Signname   string `json:"signname,omitempty"`
}

// SmsResponse 定义了接收短信回复的结构体
type SmsResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Code   string `json:"code"`
	Data   struct {
		Taskid      int `json:"taskid"`
		Validnumber int `json:"validnumber"`
		Jfts        int `json:"jfts"`
	} `json:"data"`
}

// generateMD5Sign 用于生成MD5签名
func generateMD5Sign(apikey, timestamp, apisecret string) string {
	strToHash := apikey + timestamp + apisecret
	hasher := md5.New()
	hasher.Write([]byte(strToHash))
	return hex.EncodeToString(hasher.Sum(nil))
}
