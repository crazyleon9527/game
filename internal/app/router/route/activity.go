package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterActivityRoutes(r *gin.RouterGroup, activityAPI *api.ActivityAPI) {
	activity := r.Group("/activity")
	{
		activity.POST("/get-activity-list", middleware.JWTMiddleware(), activityAPI.GetActivityList)
		activity.POST("/get-banner-list", activityAPI.GetBannerList)
		activity.POST("/get-logo-list", activityAPI.GetLogoList)
		activity.POST("/get-red-envelope", middleware.JWTMiddleware(), activityAPI.GetRedEnvelope)
		activity.POST("/join-pinduo", middleware.JWTMiddleware(), activityAPI.JoinPinduo)
		activity.POST("/get-pinduo-cash", middleware.JWTMiddleware(), activityAPI.GetPinduoCash)

		activity.POST("/admin/add-red-envelope", middleware.AdminMiddleware(), activityAPI.AddRedEnvelope)
		activity.POST("/admin/del-red-envelope", middleware.AdminMiddleware(), activityAPI.DelRedEnvelope)
	}
}
