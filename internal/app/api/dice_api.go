package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/game/dice"
	"rk-api/internal/app/ginx"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var DiceGameAPISet = wire.NewSet(wire.Struct(new(DiceGameAPI), "*"))

type DiceGameAPI struct {
	DiceGame *dice.DiceGame
}

// DiceGameGetOrderList
func (m *DiceGameAPI) DiceGameGetOrderList(ctx *gin.Context) {
	uid := ginx.Mine(ctx)

	rsp, err := m.DiceGame.GetOrderList(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// DiceGamePlaceBet 下注
func (m *DiceGameAPI) DiceGamePlaceBet(ctx *gin.Context) {
	var req *entities.DiceGamePlaceBetReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := m.DiceGame.PlaceBet(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// DiceGameChangeSeed
func (m *DiceGameAPI) DiceGameChangeSeed(ctx *gin.Context) {
	var req *entities.DiceGameChangeSeedReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := m.DiceGame.ChangeSeed(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}
