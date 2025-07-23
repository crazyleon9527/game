package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterAdminRoutes(r *gin.RouterGroup, adminAPI *api.AdminAPI) {
	admin := r.Group("/admin")
	{
		admin.POST("/check-google-code-binded", adminAPI.CheckGoogleAuthCodeBinded)
		admin.POST("/gen-auth-qr-code", adminAPI.GenAuthQRCode)
		admin.POST("/verify-google-auth-code", adminAPI.VerifyGoogleAuthCode)
		admin.POST("/call-month-backup-and-clean", middleware.OptMiddleware(), adminAPI.CallMonthBackupAndClean)
		admin.POST("/call-change-pc", middleware.AdminMiddleware(), adminAPI.CallChangePC)
	}
}
