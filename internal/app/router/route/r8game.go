package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterR8GameRoutes(r *gin.RouterGroup, r8API *api.R8API) {
	r8 := r.Group("/r8game")
	{
		r8.GET("/rich88/session_id", r8API.GetSessionIDToken)
		r8.GET("/rich88/award_activity", middleware.R88AuthMiddleware(), r8API.AwardActivity)
		r8.GET("/rich88/balance/:uid", middleware.R88AuthMiddleware(), r8API.GetBalance)
		r8.POST("/rich88/transfer", middleware.R88AuthMiddleware(), r8API.Transfer)
		r8.POST("/kick", r8API.Kick)
		r8.POST("/launch", middleware.JWTMiddleware(), r8API.Login)
	}
}
