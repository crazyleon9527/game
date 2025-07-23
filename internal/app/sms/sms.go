package sms

import "strings"

type IMessage interface {
	Send(phone string, msg string) (string, error)
}

func SendOptSms(phone string, msg string) (string, error) {
	// return SendV22tSms(phone, msg)
	normalizePhone := strings.TrimPrefix(phone, "+") //去掉前面的加号

	appKey := "c9Lcj4dy"
	appSecret := "4VDwxX32"
	apiUrl := "http://api.nxcloud.com/api/sms/mtsend"
	return NewPlantSms(appKey, appSecret, apiUrl).Send(normalizePhone, msg)
}
