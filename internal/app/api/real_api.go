package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var RealAPISet = wire.NewSet(wire.Struct(new(RealAPI), "*"))

type RealAPI struct {
	Srv *service.RealService
}

// 提交实名认证
func (c *RealAPI) CommitRealAuth(ctx *gin.Context) {

	var req entities.RealNameAuthReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	auth := entities.RealAuth{
		RealName: req.RealName,
		IDCard:   req.IDCard,
		UID:      req.UID,
	}
	err := c.Srv.CreateRealAuth(&auth)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

// 获取实名认证状态
func (c *RealAPI) GetRealAuthByUserID(ctx *gin.Context) {
	auth, err := c.Srv.GetRealAuthByUID(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, auth)
}

// 更新实名认证状态
func (c *RealAPI) UpdateRealNameAuth(ctx *gin.Context) {
	var auth entities.UpdateRealNameAuthReq
	if err := ctx.ShouldBindJSON(&auth); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.UpdateRealAuth(&auth)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}
