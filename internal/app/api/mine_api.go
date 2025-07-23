package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/game/mine"
	"rk-api/internal/app/ginx"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var MineGameAPISet = wire.NewSet(wire.Struct(new(MineGameAPI), "*"))

type MineGameAPI struct {
	MineGame *mine.MineGame
}

// MineGameGetState 获取状态
func (m *MineGameAPI) MineGameGetState(ctx *gin.Context) {
	uid := ginx.Mine(ctx)

	rsp, err := m.MineGame.GetState(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// MineGameGetOrderList
func (m *MineGameAPI) MineGameGetOrderList(ctx *gin.Context) {
	uid := ginx.Mine(ctx)

	rsp, err := m.MineGame.GetOrderList(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// MineGamePlaceBet 下注
func (m *MineGameAPI) MineGamePlaceBet(ctx *gin.Context) {
	var req *entities.MineGamePlaceBetReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := m.MineGame.PlaceBet(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// MineGameOpenPosition
func (m *MineGameAPI) MineGameOpenPosition(ctx *gin.Context) {
	var req *entities.MineGameOpenPositionReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := m.MineGame.OpenPosition(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// MineGameCashout
func (m *MineGameAPI) MineGameCashout(ctx *gin.Context) {
	uid := ginx.Mine(ctx)

	rsp, err := m.MineGame.Cashout(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// MineGameChangeSeed
func (m *MineGameAPI) MineGameChangeSeed(ctx *gin.Context) {
	var req *entities.MineGameChangeSeedReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := m.MineGame.ChangeSeed(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}
