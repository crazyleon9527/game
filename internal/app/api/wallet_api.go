package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"
	"rk-api/pkg/structure"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var WalletAPISet = wire.NewSet(wire.Struct(new(WalletAPI), "*"))

type WalletAPI struct {
	Srv *service.WalletService
}

func (c *WalletAPI) ReportFreeze(ctx *gin.Context) {
	var req entities.ReportFreezeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	fundFreeze := new(entities.FundFreeze)
	structure.Copy(req, fundFreeze)

	err := c.Srv.CreateFundFreeze(fundFreeze)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WalletAPI) UpdateWalletPassword(ctx *gin.Context) {
	var req entities.UpdateWalletPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		ginx.RespErr(ctx, errors.With("Two passwords do not match"))
		return
	}
	var uid uint = ginx.Mine(ctx)
	err := c.Srv.UpdateWalletPassword(uid, req.Password, req.NewPassword)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WalletAPI) EnableWalletPassword(ctx *gin.Context) {
	var req entities.EnableWalletPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	var uid uint = ginx.Mine(ctx)
	err := c.Srv.EnableWalletPassword(uid, req.Password)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WalletAPI) GetUserWallet(ctx *gin.Context) {
	var uid uint = ginx.Mine(ctx)
	wallet, err := c.Srv.GetWallet(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, wallet)
}
