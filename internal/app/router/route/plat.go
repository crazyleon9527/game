package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
)

func RegisterPlatRoutes(r *gin.RouterGroup, platAPI *api.PlatAPI) {
	plat := r.Group("/plat")
	{
		plat.POST("/get-platform", platAPI.GetPlatformInfo)
	}
}
