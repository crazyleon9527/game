package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
)

func RegisterHashGameRoutes(r *gin.RouterGroup, hashAPI *api.HashGameAPI) {
	hash := r.Group("/hashgame")
	{
		hash.POST("/fire-check", hashAPI.FairCheck)
	}
}
