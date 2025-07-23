package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterVerifyRoutes(r *gin.RouterGroup, verifyAPI *api.VerifyAPI) {
	verify := r.Group("/verify")
	{
		verify.POST("/send-verify-code", verifyAPI.SendVerifyCode)
		verify.POST("/admin/get-sms-verification-state", verifyAPI.GetSMSVerificationState)
		verify.POST("/admin/switch-sms-channel", middleware.AdminMiddleware(), verifyAPI.SwitchSmsChannel)
		verify.POST("/admin/toggle-sms-verification-state", middleware.AdminMiddleware(), verifyAPI.ToggleSMSVerificationState)
	}
}
