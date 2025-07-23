package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
)

func RegisterGameRoutes(r *gin.RouterGroup, gameAPI *api.GameAPI) {
	game := r.Group("/game")
	{
		game.POST("/get-game-list", gameAPI.GetGameList)
		game.POST("/search", gameAPI.SearchGame)
		game.POST("/refresh-game-list", gameAPI.RefreshGameList)
	}
}
