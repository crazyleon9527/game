package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var AdminAPISet = wire.NewSet(wire.Struct(new(AdminAPI), "*"))

type AdminAPI struct {
	Srv     *service.AdminService
	UserSrv *service.UserService
}

func (c *AdminAPI) GenAuthQRCode(ctx *gin.Context) {
	account := ctx.Query("account")

	png, err := c.Srv.GenAuthQRCode(account)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, png)
}

func (c *AdminAPI) VerifyGoogleAuthCode(ctx *gin.Context) {
	var req entities.VerifyGoogleAuthCodeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Srv.VerifyGoogleAuthCode(req.Account, req.AuthCode)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *AdminAPI) CheckGoogleAuthCodeBinded(ctx *gin.Context) {
	account := ctx.Query("account")

	ret, err := c.Srv.CheckGoogleAuthCodeBinded(account)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, ret)
}

func (c *AdminAPI) CallMonthBackupAndClean(ctx *gin.Context) {
	var req entities.MonthBackupAndCleaReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.CallMonthBackupAndClean(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *AdminAPI) CallChangePC(ctx *gin.Context) {
	var req entities.CallChangePCReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	//清除原业务员用户的RDS缓存
	err := c.UserSrv.BatchClearUserCacheByPC(req.SRC)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err = c.Srv.CallChangePC(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}
