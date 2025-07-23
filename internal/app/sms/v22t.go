package sms

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	resty "rk-api/pkg/http"
	"strconv"
	"time"
)

type V22Sms struct {
	appKey    string
	appSecret string
	apiUrl    string //http://v22t.xyz:333/api/httpSubmit
}

func NewV22Sms(appKey, appSecret, apiUrl string) IMessage {
	return &V22Sms{
		appKey:    appKey,
		appSecret: appSecret,
		apiUrl:    apiUrl,
	}
}

func (s *V22Sms) Send(phone string, msg string) (string, error) {
	// phone = "91" + phone

	// 构造请求数据
	data := map[string]string{
		"phones":  phone,
		"content": msg,
	}

	// appId := "10064"
	// appSecret := "ULqugmT_$hkB!Jam"
	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	signBytes := md5.Sum([]byte(s.appKey + s.appSecret + timestamp))
	signature := hex.EncodeToString(signBytes[:])

	// 发送请求
	response, err := resty.GetHttpClient().R().
		SetHeader("Content-Type", "application/json").
		SetHeader("appId", s.appKey).
		SetHeader("timestamp", timestamp).
		SetHeader("sign", signature).
		SetBody(data).
		Post(s.apiUrl)

	if err != nil {
		return "", err
	}

	// 检查 HTTP 响应状态码
	if response.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("received non-200 status code: %d", response.StatusCode())
	}

	return response.String(), nil
}
