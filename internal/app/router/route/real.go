package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterRealRoutes(r *gin.RouterGroup, realAPI *api.RealAPI) {
	real := r.Group("/real")
	{
		real.POST("/commit-real-auth", middleware.JWTMiddleware(), realAPI.CommitRealAuth)
		real.POST("/get-real-auth", middleware.JWTMiddleware(), realAPI.GetRealAuthByUserID)
		real.POST("/update-real-auth", middleware.JWTMiddleware(), realAPI.UpdateRealNameAuth)
	}
}
