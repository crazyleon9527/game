package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// WingoSet 注入Wingo
var WingoSet = wire.NewSet(wire.Struct(new(Wingo), "*"))

// Wingo 示例程序
type Wingo struct {
}

// @Tags Wingo
// @Summary 获取房间信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.IDReq true "params"
// @Success 200 {object} entities.WingoRoomResp
// @Router /api/wingo/get-room [post]
func (c *Wingo) GetRoom(ctx *gin.Context) {
}

// @Tags Wingo
// @Summary 获取最近历史玩家订单列表
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.WingoOrderHistoryReq true "params"
// @Success 200 {object} entities.Paginator{List=[]entities.WingoOrder}
// @Router /api/wingo/recent-order-history-list [post]
func (c *Wingo) GetRecentOrderHistoryList(ctx *gin.Context) {

}

// @Tags Wingo
// @Summary 获取最近历史期数列表
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.GetPeriodHistoryListReq true "params"
// @Success 200 {object} entities.Paginator{List=[]entities.WingoPeriod}
// @Router /api/wingo/recent-period-history-list [post]
func (c *Wingo) GetRecentPeriodHistoryList(ctx *gin.Context) {

}

// @Tags Wingo
// @Summary 同步状态
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.StateSyncReq true "params"
// @Success 200 {object} entities.WingoRoomResp
// @Router /api/wingo/state-sync [post]
func (c *Wingo) StateSync(ctx *gin.Context) {

}

// @Tags Wingo
// @Summary 创建投注单
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.WingoOrderReq true "params"
// @Success 200 {object} entities.WingoOrder
// @Router /api/wingo/create-order [post]
func (c *Wingo) CreateOrder(ctx *gin.Context) {

}

// @Tags Wingo
// @Summary 获取当前期玩家投注订单列表
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.WingoOrderReq true "params"
// @Success 200 {object} entities.Paginator{List=[]entities.WingoOrder}
// @Router /api/wingo/admin/get-period-player-orders [post]
func (c *Wingo) GetPeriodPlayerOrderList(ctx *gin.Context) {

}

// @Tags Wingo
// @Summary 获取当前期投注信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.IDReq true "params"
// @Success 200
// @Router /api/wingo/admin/get-period-bet-info [post]
func (c *Wingo) GetPeriodBetInfo(ctx *gin.Context) {

}

// @Tags Wingo
// @Summary 更新当前期中奖数字
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.UpdatePeriodReq true "params"
// @Success 200 {object} ginx.Resp
// @Router /api/wingo/admin/update-period-number [post]
func (c *Wingo) UpdatePeriodNumber(ctx *gin.Context) {

}
