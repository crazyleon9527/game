package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterJHSZGameRoutes(r *gin.RouterGroup, jhszAPI *api.JhszAPI) {
	jhsz := r.Group("/jhszgame")
	{
		jhsz.POST("/launch", middleware.JWTMiddleware(), jhszAPI.Launch)
		jhsz.POST("/transfer", jhszAPI.Transfer)
		jhsz.POST("/fetch-balance", jhszAPI.FetchBalance)
		jhsz.POST("/get-available-free-card", jhszAPI.GetAvailableFreeCard)
		jhsz.POST("/use-free-card", jhszAPI.UseFreeCard)
		jhsz.POST("/send-notification", jhszAPI.SendNotification)
		jhsz.POST("/login", jhszAPI.Login)
	}
}
