package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/game/limbo"
	"rk-api/internal/app/ginx"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var LimboGameAPISet = wire.NewSet(wire.Struct(new(LimboGameAPI), "*"))

type LimboGameAPI struct {
	LimboGame *limbo.LimboGame
}

// LimboGameGetOrderList
func (m *LimboGameAPI) LimboGameGetOrderList(ctx *gin.Context) {
	uid := ginx.Mine(ctx)

	rsp, err := m.LimboGame.GetOrderList(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// LimboGamePlaceBet 下注
func (m *LimboGameAPI) LimboGamePlaceBet(ctx *gin.Context) {
	var req *entities.LimboGamePlaceBetReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := m.LimboGame.PlaceBet(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// LimboGameChangeSeed
func (m *LimboGameAPI) LimboGameChangeSeed(ctx *gin.Context) {
	var req *entities.LimboGameChangeSeedReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := m.LimboGame.ChangeSeed(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}