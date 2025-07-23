package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterSDGameRoutes(r *gin.RouterGroup, sdAPI *api.SDGameAPI) {
	sd := r.Group("/sdgame")
	{
		sd.POST("/get-sd-game-state", sdAPI.GetSDGameState)
		sd.POST("/place-sd-bet", middleware.JWTMiddleware(), sdAPI.PlaceSDBet)
	}
}
