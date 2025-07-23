package api

import (
	"net/http"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var ZfAPISet = wire.NewSet(wire.Struct(new(ZfAPI), "*"))

type ZfAPI struct {
	Srv *service.ZfService
}

func (c *ZfAPI) Login(ctx *gin.Context) {
	var req entities.ZfGameLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	data, err := c.Srv.Login(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *ZfAPI) Launch(ctx *gin.Context) {
	var req entities.ZfGameLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	data, err := c.Srv.Launch(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *ZfAPI) Register(ctx *gin.Context) {

	data, err := c.Srv.Register(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *ZfAPI) Kick(ctx *gin.Context) {
	var req entities.ZfKickReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	data, err := c.Srv.Kick(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *ZfAPI) QueryPlayer(ctx *gin.Context) {

	data, err := c.Srv.QueryPlayer(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *ZfAPI) Refund(ctx *gin.Context) {
	var req entities.ZfRefund
	// logger.Error("------------Refund----------------")
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	data := c.Srv.Refund(&req)
	// logger.ZError("------------Refund----------------", zap.Any("data", data), zap.Any("req", req))
	ctx.JSON(http.StatusOK, data)
}

func (c *ZfAPI) Bet(ctx *gin.Context) {
	var req entities.ZfBetReq
	// logger.Error("------------Bet----------------")
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	data := c.Srv.Bet(&req)
	// logger.ZError("------------Bet----------------", zap.Any("data", data), zap.Any("req", req))
	ctx.JSON(http.StatusOK, data)
}

func (c *ZfAPI) Payout(ctx *gin.Context) {
	var req entities.ZfPayoutReq
	// logger.Error("------------Payout----------------")
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	data := c.Srv.Payout(&req)
	// logger.ZError("------------Payout----------------", zap.Any("data", data), zap.Any("req", req))
	ctx.JSON(http.StatusOK, data)
}

func (c *ZfAPI) FetchBalance(ctx *gin.Context) {
	var req entities.ZfBalanceReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	data := c.Srv.FetchBalance(&req)
	// logger.ZError("------------FetchBalance----------------", zap.Any("data", data), zap.Any("req", req))
	ctx.JSON(http.StatusOK, data)
}

func (c *ZfAPI) Settle(ctx *gin.Context) {
	var req entities.ZfSettleReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	data := c.Srv.Settle(&req)
	ctx.JSON(http.StatusOK, data)
}
