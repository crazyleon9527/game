package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var AgentAPISet = wire.NewSet(wire.Struct(new(AgentAPI), "*"))

type AgentAPI struct {
	Srv *service.AgentService
}

func (c *AgentAPI) FixLevel1InviteCount(ctx *gin.Context) {
	list, err := c.Srv.FixLevel1InviteCount()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, list)
}

func (c *AgentAPI) GetGameRebateReceiptList(ctx *gin.Context) {
	var req entities.GetGameRebateReceiptListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.GetGameRebateReceiptList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *AgentAPI) GetReturnCash(ctx *gin.Context) {
	// var req entities.IDReq
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	ginx.RespErr(ctx, err)
	// 	return
	// }

	err := c.Srv.FinalizeGameCashReturn(ginx.Mine(ctx))

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *AgentAPI) GetMonthRechargeCashAlreadyReturn(ctx *gin.Context) {

	num, err := c.Srv.GetMonthRechargeCashAlreadyReturn(ginx.Mine(ctx))

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, num)
}

func (c *AgentAPI) GetPromotionProfit(ctx *gin.Context) {

	promotionProfit, err := c.Srv.GetPromotionProfit(ginx.Mine(ctx))

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, promotionProfit)
}

func (c *AgentAPI) GetPromotionList(ctx *gin.Context) {

	var req entities.GetPromotionListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	err := c.Srv.GetPromotionList(&req)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *AgentAPI) GetPromotionLink(ctx *gin.Context) {

	info, err := c.Srv.GetPromotionLink(ginx.Mine(ctx))

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, info)
}

func (c *AgentAPI) FinalizeRechargeCashReturn(ctx *gin.Context) {

	var req entities.FinalizeRechargeReturnReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.IP = ctx.ClientIP()

	err := c.Srv.FinalizeRechargeCashReturn(&req)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *AgentAPI) FixInviteRelation(ctx *gin.Context) {
	var req entities.FixInviteRelationReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.IP = ctx.ClientIP()
	err := c.Srv.FixInviteRelation(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}
