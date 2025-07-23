package api

import (
	"rk-api/internal/app/entities"
	game "rk-api/internal/app/game/rg"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var NineAPISet = wire.NewSet(wire.Struct(new(NineAPI), "*"))

type NineAPI struct {
	Srv *service.NineService
	// Nine *game.Nine
	Nine game.INine
}

func (c *NineAPI) GetRoom(ctx *gin.Context) {
	var req entities.GetRoomReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	roomInfo := c.Nine.GetInfo(req.BetType)
	ginx.RespSucc(ctx, roomInfo)
}

func (c *NineAPI) GetRecentOrderHistoryList(ctx *gin.Context) {

	var req entities.NineOrderHistoryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	err := c.Nine.GetRecentOrderHistoryList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *NineAPI) GetRecentPeriodHistoryList(ctx *gin.Context) {

	var req entities.GetPeriodHistoryListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Nine.GetRecentPeriodHistoryList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *NineAPI) StateSync(ctx *gin.Context) {

	var req entities.StateSyncReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	roomInfoResp := c.Nine.StateSync(&req)
	ginx.RespSucc(ctx, roomInfoResp)

}

func (c *NineAPI) CreateOrder(ctx *gin.Context) {

	var req entities.NineOrderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	order, err := c.Nine.CreateOrder(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, order)
}

func (c *NineAPI) SimulateSettleOrders(ctx *gin.Context) {

	var req entities.SimulateSettleOrdersReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.SimulateSettleNineOrders(req.PeriodID, req.BetType)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *NineAPI) GetTrendInfo(ctx *gin.Context) {

	var req entities.WingoTrendReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	info, err := c.Srv.GetTodayTrend(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, info)
}

/********************************************************admin***************************************************************************/

func (c *NineAPI) GetPeriodPlayerOrderList(ctx *gin.Context) {

	var req entities.OrderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	result := c.Nine.GetPeriodPlayerOrderList(&req)

	ginx.RespSucc(ctx, result)
}

func (c *NineAPI) GetPeriodBetInfo(ctx *gin.Context) {

	var req entities.GetPeriodBetInfoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	order := c.Nine.GetPeriodBetInfo(&req)

	ginx.RespSucc(ctx, order)
}

func (c *NineAPI) ChangePeriodNumber(ctx *gin.Context) {

	var req entities.UpdatePeriodReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Nine.UpdatePeriodNumber(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *NineAPI) GetTodayPeriodList(ctx *gin.Context) {
	var req entities.GetPeriodListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Srv.GetTodayPeriodList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *NineAPI) GetPeriodInfo(ctx *gin.Context) {
	var req entities.GetPeriodReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	info := c.Nine.GetPeriodInfo(&req)
	ginx.RespSucc(ctx, info)
}
