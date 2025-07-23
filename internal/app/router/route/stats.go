package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterStatsRoutes(r *gin.RouterGroup, statsAPI *api.StatsAPI) {
	stats := r.Group("/stats")
	{
		stats.POST("/get-category-stats-list", middleware.JWTMiddleware(), statsAPI.GetCategoryStatsList)
		stats.POST("/get-game-stats", middleware.JWTMiddleware(), statsAPI.GetGameStats)
		stats.POST("/get-game-record-list", middleware.JWTMiddleware(), statsAPI.GetGameRecordList)
		stats.POST("/get-gamer-daily-stats-list", middleware.JWTMiddleware(), statsAPI.GetGamerDailyStatsList)
		stats.POST("/get-gamer-profit-leaderboard", middleware.JWTMiddleware(), statsAPI.GetUserProfitLeaderboard)
		stats.POST("/get-profit-leaderboard", statsAPI.GetProfitLeaderboard)
	}
}
