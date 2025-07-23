package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterWingoRoutes(r *gin.RouterGroup, wingoAPI *api.WingoAPI) {
	wingo := r.Group("/wingo")
	{
		wingo.POST("/get-room", wingoAPI.GetRoom)
		wingo.POST("/state-sync", wingoAPI.StateSync)
		wingo.POST("/get-trend-info", middleware.JWTMiddleware(), wingoAPI.GetTrendInfo)
		wingo.POST("/recent-period-history-list", middleware.JWTMiddleware(), wingoAPI.GetRecentPeriodHistoryList)
		wingo.POST("/recent-order-history-list", middleware.JWTMiddleware(), wingoAPI.GetRecentOrderHistoryList)
		wingo.POST("/create-order", middleware.JWTMiddleware(), wingoAPI.CreateOrder)
		wingo.POST("/simulate-settle-orders", wingoAPI.SimulateSettleOrders)

		wingo.POST("/admin/get-period-player-orders", middleware.AdminMiddleware(), wingoAPI.GetPeriodPlayerOrderList)
		wingo.POST("/admin/get-period-bet-info", middleware.AdminMiddleware(), wingoAPI.GetPeriodBetInfo)
		wingo.POST("/admin/change-period-number", middleware.AdminMiddleware(), wingoAPI.ChangePeriodNumber)
		wingo.POST("/admin/get-today-period-list", middleware.AdminMiddleware(), wingoAPI.GetTodayPeriodList)
		wingo.POST("/admin/get-period-info", middleware.AdminMiddleware(), wingoAPI.GetPeriodInfo)
		wingo.POST("/admin/update-room-limit", middleware.AdminMiddleware(), wingoAPI.UpdateRoomLimit)
	}
}
