package api

import (
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/game/crash"
	"rk-api/internal/app/game/dice"
	"rk-api/internal/app/game/mine"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var HashGameAPISet = wire.NewSet(wire.Struct(new(HashGameAPI), "*"))

type HashGameAPI struct {
	Srv *service.HashGameService
}

var fairCheckMap = map[string]func(*entities.FairCheckReq) (*entities.FairCheckRsp, error){
	constant.GameNameCrash: crash.FairCheck,
	constant.GameNameMine:  mine.FairCheck,
	constant.GameNameDice:  dice.FairCheck,
}

func (h *HashGameAPI) FairCheck(ctx *gin.Context) {
	var req *entities.FairCheckReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	check, ok := fairCheckMap[req.Game]
	if !ok {
		ginx.RespErr(ctx, errors.With("game not found"))
		return
	}
	rsp, err := check(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}
