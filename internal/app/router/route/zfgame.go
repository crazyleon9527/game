package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterZFGameRoutes(r *gin.RouterGroup, zfAPI *api.ZfAPI) {
	zfgame := r.Group("/zfgame")
	{
		zfgame.POST("/launch", middleware.JWTMiddleware(), zfAPI.Launch)
		zfgame.POST("/player_summary", zfAPI.FetchBalance)
		zfgame.POST("/settle", zfAPI.Settle)
		zfgame.POST("/kick", zfAPI.Kick)
		zfgame.POST("/refund", zfAPI.Refund)
		zfgame.POST("/bet", zfAPI.Bet)
		zfgame.POST("/payout", zfAPI.Payout)
	}
}
