package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/game/hash"
	"rk-api/internal/app/ginx"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/spf13/cast"
)

var SDGameAPISet = wire.NewSet(wire.Struct(new(SDGameAPI), "*"))

type SDGameAPI struct {
	GameManage *hash.GameManage
}

func (c *SDGameAPI) PlaceSDBet(ctx *gin.Context) {
	var req *entities.HashSDBetRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	room, err := c.GameManage.Get(hash.GameStrategyTypeSingleDouble, hash.RoomType(req.GetBetType()))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err = room.HandleBet(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *SDGameAPI) GetSDGameState(ctx *gin.Context) {
	rawBetType := ctx.DefaultQuery("betType", "1")
	betType := cast.ToUint8(rawBetType)
	room, err := c.GameManage.Get(hash.GameStrategyTypeSingleDouble, hash.RoomType(betType))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	stat := room.GetGameState()
	ginx.RespSucc(ctx, stat)
}
