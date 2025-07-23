package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var VerifyAPISet = wire.NewSet(wire.Struct(new(VerifyAPI), "*"))

type VerifyAPI struct {
	Srv *service.VerifyService
}

func (c *VerifyAPI) SendVerifyCode(ctx *gin.Context) {

	var req entities.VerifyCodeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	verifyCode := new(entities.VerifyCode)
	verifyCode.Target = req.Target
	verifyCode.VerificationType = req.VerificationType
	verifyCode.BusinessType = req.BusinessType

	err := c.Srv.SendVerifyCode(verifyCode)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *VerifyAPI) SwitchSmsChannel(ctx *gin.Context) {
	var req entities.SwitchChannelReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.SwitchSmsChannel(req.Channel)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *VerifyAPI) ToggleSMSVerificationState(ctx *gin.Context) {
	var req entities.ToggleSMSVerificationStateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.ToggleSMSVerificationState(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *VerifyAPI) GetSMSVerificationState(ctx *gin.Context) {
	ret := c.Srv.GetSMSVerificationState() //bool
	ginx.RespSucc(ctx, &entities.SMSVerificationStateResp{
		Disabled: ret,
	})
}
