package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterNineRoutes(r *gin.RouterGroup, nineAPI *api.NineAPI) {
	nine := r.Group("/nine")
	{
		nine.POST("/get-room", nineAPI.GetRoom)
		nine.POST("/state-sync", nineAPI.StateSync)
		nine.POST("/recent-period-history-list", middleware.JWTMiddleware(), nineAPI.GetRecentPeriodHistoryList)
		nine.POST("/recent-order-history-list", middleware.JWTMiddleware(), nineAPI.GetRecentOrderHistoryList)
		nine.POST("/create-order", middleware.JWTMiddleware(), nineAPI.CreateOrder)

		nine.POST("/simulate-settle-orders", nineAPI.SimulateSettleOrders)

		nine.POST("/admin/get-period-player-orders", middleware.AdminMiddleware(), nineAPI.GetPeriodPlayerOrderList)
		nine.POST("/admin/get-period-bet-info", middleware.AdminMiddleware(), nineAPI.GetPeriodBetInfo)
		nine.POST("/admin/change-period-number", middleware.AdminMiddleware(), nineAPI.ChangePeriodNumber)
		nine.POST("/admin/get-today-period-list", middleware.AdminMiddleware(), nineAPI.GetTodayPeriodList)
		nine.POST("/admin/get-period-info", middleware.AdminMiddleware(), nineAPI.GetPeriodInfo)
	}
}
