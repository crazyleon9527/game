package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterDiceGameRoutes(r *gin.RouterGroup, diceAPI *api.DiceGameAPI) {
	dice := r.Group("/dicegame")
	{
		dice.POST("/dice-game-get-order-list", middleware.JWTMiddleware(), diceAPI.DiceGameGetOrderList)
		dice.POST("/dice-game-place-bet", middleware.JWTMiddleware(), diceAPI.DiceGamePlaceBet)
		dice.POST("/dice-game-change-seed", middleware.JWTMiddleware(), diceAPI.DiceGameChangeSeed)
	}
}
