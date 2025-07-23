package mock

import (
	"github.com/google/wire"
)

// NineSet 注入Nine
var NineSet = wire.NewSet(wire.Struct(new(Nine), "*"))

// Nine 示例程序
type Nine struct {
}

// // @Tags Nine
// // @Summary 获取房间信息
// // @Accept  json
// // @Produce  json
// // @Security ApiKeyAuth
// // @Param req body entities.IDReq true "params"
// // @Success 200
// // @Router /api/nine/get-room [post]
// func (c *Nine) GetRoom(ctx *gin.Context) {
// }

// // @Tags Nine
// // @Summary 获取最近历史玩家订单列表
// // @Accept  json
// // @Produce  json
// // @Security ApiKeyAuth
// // @Param req body entities.NineOrderHistoryReq true "params"
// // @Success 200 {object} entities.Paginator{List=[]entities.NineOrder}
// // @Router /api/nine/recent-order-history-list [post]
// func (c *Nine) GetRecentOrderHistoryList(ctx *gin.Context) {

// }

// // @Tags Nine
// // @Summary 获取最近历史期数列表
// // @Accept  json
// // @Produce  json
// // @Security ApiKeyAuth
// // @Param req body entities.GetPeriodHistoryListReq true "params"
// // @Success 200 {object} entities.Paginator{List=[]entities.NinePeriod}
// // @Router /api/nine/recent-period-history-list [post]
// func (c *Nine) GetRecentPeriodHistoryList(ctx *gin.Context) {

// }

// // @Tags Nine
// // @Summary 同步状态
// // @Accept  json
// // @Produce  json
// // @Security ApiKeyAuth
// // @Param req body entities.StateSyncReq true "params"
// // @Success 200 {object} entities.NineRoomResp
// // @Router /api/nine/state-sync [post]
// func (c *Nine) StateSync(ctx *gin.Context) {

// }

// // @Tags Nine
// // @Summary 创建投注单
// // @Accept  json
// // @Produce  json
// // @Security ApiKeyAuth
// // @Param req body entities.NineOrderReq true "params"
// // @Success 200 {object} entities.NineOrder
// // @Router /api/nine/create-order [post]
// func (c *Nine) CreateOrder(ctx *gin.Context) {

// }

// // @Tags Nine
// // @Summary 获取当前期玩家投注订单列表
// // @Accept  json
// // @Produce  json
// // @Security ApiKeyAuth
// // @Param req body entities.NineOrderReq true "params"
// // @Success 200 {object} entities.Paginator{List=[]entities.NineOrder}
// // @Router /api/nine/admin/get-period-player-orders [post]
// func (c *Nine) GetPeriodPlayerOrderList(ctx *gin.Context) {

// }

// // @Tags Nine
// // @Summary 获取当前期投注信息
// // @Accept  json
// // @Produce  json
// // @Security ApiKeyAuth
// // @Param req body entities.IDReq true "params"
// // @Success 200 {object}
// // @Router /api/nine/admin/get-period-bet-info [post]
// func (c *Nine) GetPeriodBetInfo(ctx *gin.Context) {

// }

// // @Tags Nine
// // @Summary 更新当前期中奖数字
// // @Accept  json
// // @Produce  json
// // @Security ApiKeyAuth
// // @Param req body entities.UpdatePeriodReq true "params"
// // @Success 200 {object} ginx.Resp
// // @Router /api/nine/admin/update-period-number [post]
// func (c *Nine) UpdatePeriodNumber(ctx *gin.Context) {

// }
