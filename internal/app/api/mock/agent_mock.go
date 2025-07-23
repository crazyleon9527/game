package mock

import (
	// "rk-api/internal/app/entities"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// var req entities.IDReq

// AgentSet 注入Agent
var AgentSet = wire.NewSet(wire.Struct(new(Agent), "*"))

// Agent 示例程序
type Agent struct {
}

// @Tags Agent
// @Summary 获取返利
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} ginx.Resp{}
// @Router /api/agent/get-return-cash [post]
func (a *Agent) GetReturnCash(c *gin.Context) {
}

// @Tags Agent
// @Summary 获取推广列表
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.GetPromotionListReq true "params"
// @Success 200 {object} entities.Paginator{List=[]entities.HallInviteRelation}
// @Router /api/agent/get-promotion-list [post]
func (a *Agent) GetPromotionList(c *gin.Context) {
}

// @Tags Agent
// @Summary 获取推广盈利信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entities.PromotionProfit{}
// @Router /api/agent/get-promotion-profit [post]
func (a *Agent) GetPromotionProfit(c *gin.Context) {
}

// @Tags Agent
// @Summary 获取推广链接
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} ginx.Resp{}
// @Router /api/agent/get-promotion-link [post]
func (a *Agent) GetPromotionLink(c *gin.Context) {
}

// @Summary 获取领取的佣金记录
// @Description 获取领取的佣金记录
// @Tags Agent
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.GetGameRebateReceiptListReq true "查询条件"
// @Success 200 {array} entities.GameRebateReceipt
// @Router /api/agent/get-game-rabate-receipt-list [post]
func (c *Agent) GetGameRebateReceiptList(ctx *gin.Context) {

}
