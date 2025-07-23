package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterAuthRoutes(r *gin.RouterGroup, authAPI *api.AuthAPI) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", authAPI.RegisterUser)
		auth.POST("/login", authAPI.Login)
		auth.POST("/mobile-login", authAPI.MobileLogin)
		auth.POST("/verify-credentials", authAPI.VerifyCredentials)
		auth.POST("/logout", authAPI.Logout)
		auth.POST("/update-password", authAPI.ChangePassword)
		auth.POST("/reset-password", authAPI.ResetPassword)

		auth.POST("/admin/register-auth-user", middleware.AdminMiddleware(), authAPI.ReigsterAuthUser)
		auth.GET("/ping", authAPI.Ping)
	}
}
