package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterRechargeRoutes(r *gin.RouterGroup, rechargeAPI *api.RechargeAPI) {
	recharge := r.Group("/recharge")
	{
		recharge.POST("/get-recharge-order-list", middleware.JWTMiddleware(), rechargeAPI.GetRechargeOrderList)
		recharge.POST("/get-recharge-config", middleware.JWTMiddleware(), rechargeAPI.GetRechargeConfig)
		recharge.POST("/get-recharge-url", middleware.JWTMiddleware(), rechargeAPI.GetRechargeUrlInfo)

		// 回调接口
		recharge.POST("/callback/poly", rechargeAPI.POLYCallback)
		recharge.POST("/callback/kb", rechargeAPI.KBCallback)
		recharge.POST("/callback/tk", rechargeAPI.TKCallback)
		recharge.POST("/callback/at", rechargeAPI.ATCallback)
		recharge.POST("/callback/go", rechargeAPI.GOCallback)
		recharge.POST("/callback/cow", rechargeAPI.COWCallback)
		recharge.POST("/callback/ant", rechargeAPI.ANTCallback)
		recharge.POST("/callback/dy", rechargeAPI.DYCallback)
		recharge.POST("/callback/gaga", rechargeAPI.GAGACallback)
	}
}
