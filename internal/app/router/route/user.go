package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterUserRoutes(r *gin.RouterGroup, userAPI *api.UserAPI) {
	user := r.Group("/user")
	{
		user.POST("/search-user", userAPI.SearchUser)
		user.POST("/get-user-info", middleware.JWTMiddleware(), userAPI.GetUserInfo)
		user.POST("/get-bet-user-info", middleware.JWTMiddleware(), userAPI.GetUserInfo)
		user.POST("/get-customer", middleware.JWTMiddleware(), userAPI.GetCustomer)
		user.POST("/edit-nickname", middleware.JWTMiddleware(), userAPI.EditNickname)
		user.POST("/edit-avatar", middleware.JWTMiddleware(), userAPI.EditAvatar)
		user.POST("/bind-telegram", middleware.JWTMiddleware(), userAPI.BindTelegram)
		user.POST("/bind-email", middleware.JWTMiddleware(), userAPI.BindEmail)
		user.POST("/upload-avatar", middleware.JWTMiddleware(), userAPI.UploadAvatar)
		user.POST("/admin/edit-user", middleware.AdminMiddleware(), userAPI.EditUserInfo)
		user.POST("/admin/clear-user-cache", userAPI.ClearUserCache)
	}
}
