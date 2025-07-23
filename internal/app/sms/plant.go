package sms

import (
	"encoding/json"
	"fmt"
	"net/url"
	"rk-api/pkg/http"
)

type NxCloudResponse struct {
	MessageID string `json:"messageid"`
}

type PlantSms struct {
	appKey    string
	appSecret string
	apiUrl    string //http://api.nxcloud.com/api/sms/mtsend
}

func NewPlantSms(appKey, appSecret, apiUrl string) IMessage {
	return &PlantSms{
		appKey:    appKey,
		appSecret: appSecret,
		apiUrl:    apiUrl,
	}
}

func (s *PlantSms) Send(phone, content string) (string, error) {
	client := http.GetHttpClient()
	// URLEncode the content
	encodedContent := url.QueryEscape(content)
	// Construct the data for the POST body
	data := fmt.Sprintf("appkey=%s&secretkey=%s&phone=%s&content=%s", s.appKey, s.appSecret, phone, encodedContent)

	// Perform the POST request
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8").
		SetBody(data).
		Post(s.apiUrl)

	if err != nil {
		return "", err
	}

	// Parse the JSON response
	var result NxCloudResponse
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return "", err
	}
	// Return the message ID
	return result.MessageID, nil
}
