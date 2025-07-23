package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"
	"rk-api/internal/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var JhszAPISet = wire.NewSet(wire.Struct(new(JhszAPI), "*"))

type JhszAPI struct {
	Srv *service.JhszService
}

func (c *JhszAPI) Launch(ctx *gin.Context) {
	var req entities.JhszGameLoginReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	req.Ip = ctx.ClientIP()
	userAgent := ctx.Request.Header.Get("User-Agent")
	req.DeviceOS = utils.GetDeviceOS(userAgent)

	deviceId, err := ctx.Cookie("deviceId")
	if err != nil {
		// 如果不存在，则生成一个新的 deviceId
		deviceId = ginx.SetDeviceID(ctx)
	}
	req.DeviceId = deviceId
	if req.Language == "" {
		req.Language = ginx.GetPreferredLanguage(ctx)
	}

	req.UID = ginx.Mine(ctx)
	data, err := c.Srv.Launch(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *JhszAPI) FetchBalance(ctx *gin.Context) {
	var req entities.JhszBalanceReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	data, err := c.Srv.FetchWallet(&req)
	// logger.ZError("------------FetchBalance----------------", zap.Any("data", data), zap.Any("req", req))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *JhszAPI) Transfer(ctx *gin.Context) {
	var req entities.JhszTransferReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	data, err := c.Srv.Transfer(&req)
	// logger.ZInfo("------------Transfer-", zap.Any("data", data), zap.Any("req", req))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *JhszAPI) GetAvailableFreeCard(ctx *gin.Context) {
	var req entities.GetAvailableFreeCardReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	data, err := c.Srv.GetAvailableFreeCard(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *JhszAPI) UseFreeCard(ctx *gin.Context) {
	var req entities.UseFreeCardReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	data, err := c.Srv.UseFreeCard(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *JhszAPI) SendNotification(ctx *gin.Context) {
	var req entities.JhszNotificationReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	c.Srv.SendNotification(&req)
	ginx.RespSucc(ctx, nil)
}

func (c *JhszAPI) Login(ctx *gin.Context) {
	var credentials entities.PlatLoginReq
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	credentials.LoginIP = ctx.ClientIP()

	user, err := c.Srv.Login(&credentials)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, user)
}
