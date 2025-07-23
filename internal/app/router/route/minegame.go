package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterMineGameRoutes(r *gin.RouterGroup, mineAPI *api.MineGameAPI) {
	mine := r.Group("/minegame")
	{
		mine.POST("/mine-game-get-state", middleware.JWTMiddleware(), mineAPI.MineGameGetState)
		mine.POST("/mine-game-get-order-list", middleware.JWTMiddleware(), mineAPI.MineGameGetOrderList)
		mine.POST("/mine-game-place-bet", middleware.JWTMiddleware(), mineAPI.MineGamePlaceBet)
		mine.POST("/mine-game-open-position", middleware.JWTMiddleware(), mineAPI.MineGameOpenPosition)
		mine.POST("/mine-game-cashout", middleware.JWTMiddleware(), mineAPI.MineGameCashout)
		mine.POST("/mine-game-change-seed", middleware.JWTMiddleware(), mineAPI.MineGameChangeSeed)
	}
}
