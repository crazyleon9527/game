package test

import (
	"fmt"
	"reflect"
	"rk-api/internal/app/sms"
	"testing"
	"time"
)

type User struct {
	Name    string
	Balance uint
	gender  uint
	Age     uint
	Email   string
	Date    time.Time
}

func ToMap(u User) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(u)
	t := reflect.TypeOf(u)

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)

		if fieldValue.Kind() == reflect.String && fieldValue.String() != "" {
			result[fieldType.Name] = fieldValue.String()
		} else if fieldValue.Kind() == reflect.Int && fieldValue.Int() != 0 {
			result[fieldType.Name] = fieldValue.Int()
		}
	}

	return result
}

func TestCopy(t *testing.T) {

	// user := &User{
	// 	Name:    "luoyuxiang",
	// 	Balance: 0,
	// 	gender:  1001,
	// 	Age:     4443,
	// }

	// toValue := make(map[string]interface{}, 0)

	// toValue2 := make(map[string]interface{})

	// structure.Copy(user, &toValue2)

	// copier.CopyWithOption(&toValue, user, copier.Option{DeepCopy: true})
	// t.Log(toValue["Name"])
	// t.Logf("%v", toValue)
	// t.Logf("%v", toValue2)

	// u := User{Name: "John", Email: "", Age: 2, Date: time.Now()}

	// t.Log(ToMap(u))

	// r := make(map[string]interface{})

	// copier.Copy(r, u)

	// t.Log(r)

	// s := service.ProvideZfService(nil, nil)

	// var ss string = `{"unique_id":"LkTxngJ57p1hZLrev69Ek","timestamp":1708580180,"merchant_code":"dxicpnFJLpuxBynTpCDmj","sign":"7e9710b6441ff166b318d16d1fb4f1598e92854e","username":"800000","game_code":"crash","bet_id":1000654060,"round_id":2603663,"amount":1}`

	// // var ss string = `{"merchant_code":"dxicpnFJLpuxBynTpCDmj","sign":"ab9e946d21e7b81c86ef1dfbc2a27507b7942368","timestamp":1708571270,"unique_id":"XWxfM8AKUCHShhO4yEC-V","username":"800000"}`

	// var data map[string]interface{}
	// cjson.Cjson.UnmarshalFromString(ss, &data)

	// // delete(data, "unique_id")
	// tt := s.ValidateSignature(data)

	// // 91bef33b208ba98ac2d889b9c4e996651ca58146
	// // 91bef33b208ba98ac2d889b9c4e996651ca58146
	// fmt.Println(tt)

	// order := entities.WingoOrder{
	// 	Rate:      5,
	// 	BetAmount: 10,
	// }
	// order.CalculateFee()

	// fmt.Println(order.Delivery, order.Fee)

	ret, err := sms.SendOptSms("8618576410105", fmt.Sprintf("Dear user, your verfiy code is %s.", "4321"))

	fmt.Println(ret, err)
}
