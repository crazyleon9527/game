package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterFlowRoutes(r *gin.RouterGroup, flowAPI *api.FlowAPI) {
	flow := r.Group("/flow")
	{
		flow.POST("/get-flow-list", middleware.JWTMiddleware(), flowAPI.GetFlowList)
	}
}
