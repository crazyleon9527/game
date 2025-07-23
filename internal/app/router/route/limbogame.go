package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterLimboGameRoutes(r *gin.RouterGroup, limboAPI *api.LimboGameAPI) {
	limbo := r.Group("/limbogame")
	{
		limbo.POST("/limbo-game-get-order-list", middleware.JWTMiddleware(), limboAPI.LimboGameGetOrderList)
		limbo.POST("/limbo-game-place-bet", middleware.JWTMiddleware(), limboAPI.LimboGamePlaceBet)
		limbo.POST("/limbo-game-change-seed", middleware.JWTMiddleware(), limboAPI.LimboGameChangeSeed)
	}
}
