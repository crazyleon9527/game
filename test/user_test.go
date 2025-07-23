package test

import (
	"rk-api/internal/app/entities"
	"rk-api/pkg/cjson"
	"testing"
)

func TestUser(t *testing.T) {

	user := entities.User{}
	// user.AddBalance(23.34)
	// user.AddBalance(27.7)
	// user.SubBalance(10)
	// user.SubBalance(-20)
	// user.AddRechargeAll(21.2)
	// user.AddWithdrawAll(334.12345678)

	u, _ := cjson.Cjson.MarshalToString(user)

	t.Log(u)

}
