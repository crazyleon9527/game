package test

import (
	"rk-api/internal/app/sms"
	"strings"
	"testing"
	"time"

	"github.com/orca-zhang/ecache"
)

func TestSMS(t *testing.T) {

	lr := ecache.NewLRUCache(1, 2, 5*time.Second) //初始化缓存

	lr.Put("name", 123)

	t.Log(lr.Get("name"))

	time.Sleep(6 * time.Second)
	t.Log(lr.Get("name"))
	// t.Log(SendOptSms("+918535010774", "Cheetah：Your code is 123123"))
}

func SendOptSms(phone string, msg string) (string, error) {
	// // return SendV22tSms(phone, msg)
	normalizePhone := strings.TrimPrefix(phone, "+") //去掉前面的加号
	appKey := "dn+citpRTz2fUBzPY+8g0w=="
	appSecret := "d5b98058c4a44608a7db9a5cad5de002"
	apiUrl := "https://api.i51sms.com/outauth/verifCodeSend"
	return sms.NewI51Sms(appKey, appSecret, apiUrl).Send(normalizePhone, msg)

}
