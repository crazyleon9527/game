package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterAgentRoutes(r *gin.RouterGroup, agentAPI *api.AgentAPI) {
	agent := r.Group("/agent")
	{
		agent.POST("/get-return-cash", middleware.JWTMiddleware(), agentAPI.GetReturnCash)
		agent.POST("/get-month-recharge-return", middleware.JWTMiddleware(), agentAPI.GetMonthRechargeCashAlreadyReturn)
		agent.POST("/get-promotion-profit", middleware.JWTMiddleware(), agentAPI.GetPromotionProfit)
		agent.POST("/get-promotion-list", middleware.JWTMiddleware(), agentAPI.GetPromotionList)
		agent.POST("/get-promotion-link", middleware.JWTMiddleware(), agentAPI.GetPromotionLink)
		agent.POST("/get-game-rabate-receipt-list", middleware.JWTMiddleware(), agentAPI.GetGameRebateReceiptList)

		agent.POST("/admin/finalize-recharge-return", middleware.AdminMiddleware(), agentAPI.FinalizeRechargeCashReturn)
		agent.POST("/admin/fix-invite-relation", middleware.OptMiddleware(), agentAPI.FixInviteRelation)
		agent.POST("/admin/fix-level1-invite-count", agentAPI.FixLevel1InviteCount)
	}
}
