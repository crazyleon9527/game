package api

import (
	"rk-api/internal/app/entities"
	game "rk-api/internal/app/game/rg"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var WingoAPISet = wire.NewSet(wire.Struct(new(WingoAPI), "*"))

type WingoAPI struct {
	Srv   *service.WingoService
	Wingo game.IWingo
}

func (c *WingoAPI) GetRoom(ctx *gin.Context) {
	var req entities.GetRoomReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	roomInfo := c.Wingo.GetInfo(req.BetType)
	ginx.RespSucc(ctx, roomInfo)
}

func (c *WingoAPI) GetRecentOrderHistoryList(ctx *gin.Context) {

	var req entities.WingoOrderHistoryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	err := c.Wingo.GetRecentOrderHistoryList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *WingoAPI) GetRecentPeriodHistoryList(ctx *gin.Context) {

	var req entities.GetPeriodHistoryListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Wingo.GetRecentPeriodHistoryList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *WingoAPI) StateSync(ctx *gin.Context) {

	var req entities.StateSyncReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	roomInfoResp := c.Wingo.StateSync(&req)
	ginx.RespSucc(ctx, roomInfoResp)

}

func (c *WingoAPI) CreateOrder(ctx *gin.Context) {

	var req entities.WingoOrderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	order, err := c.Wingo.CreateOrder(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, order)
}

func (c *WingoAPI) SimulateSettleOrders(ctx *gin.Context) {

	var req entities.SimulateSettleOrdersReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.SimulateSettleWingoOrders(req.PeriodID, req.BetType)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WingoAPI) GetTrendInfo(ctx *gin.Context) {

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

func (c *WingoAPI) GetPeriodPlayerOrderList(ctx *gin.Context) {

	var req entities.OrderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	result := c.Wingo.GetPeriodPlayerOrderList(&req)

	ginx.RespSucc(ctx, result)
}

func (c *WingoAPI) GetPeriodBetInfo(ctx *gin.Context) {

	var req entities.GetPeriodBetInfoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	order := c.Wingo.GetPeriodBetInfo(&req)

	ginx.RespSucc(ctx, order)
}

func (c *WingoAPI) ChangePeriodNumber(ctx *gin.Context) {

	var req entities.UpdatePeriodReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Wingo.UpdatePeriodNumber(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WingoAPI) GetTodayPeriodList(ctx *gin.Context) {
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
func (c *WingoAPI) GetPeriodInfo(ctx *gin.Context) {
	var req entities.GetPeriodReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	info := c.Wingo.GetPeriodInfo(&req)

	ginx.RespSucc(ctx, info)
}

func (c *WingoAPI) UpdateRoomLimit(ctx *gin.Context) {
	var req entities.UpdateRoomLimitReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Srv.UpdateRoomLimit(&req)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, nil)
}
