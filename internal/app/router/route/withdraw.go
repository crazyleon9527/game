package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterWithdrawRoutes(r *gin.RouterGroup, withdrawAPI *api.WithdrawAPI) {
	withdraw := r.Group("/withdraw")
	{
		withdraw.POST("/get-withdraw-detail", middleware.JWTMiddleware(), withdrawAPI.GetWithdrawDetail)
		withdraw.POST("/get-withdraw-card-list", middleware.JWTMiddleware(), withdrawAPI.GetithdrawCardList)
		withdraw.POST("/get-withdraw-record-list", middleware.JWTMiddleware(), withdrawAPI.GetHallWidthdrawRecordList)
		withdraw.POST("/add-withdraw-card", middleware.JWTMiddleware(), withdrawAPI.AddWithdrawCard)
		withdraw.POST("/del-withdraw-card", middleware.JWTMiddleware(), withdrawAPI.DelWithdrawCard)
		withdraw.POST("/select-withdraw-card", middleware.JWTMiddleware(), withdrawAPI.SelectWithdrawCard)
		withdraw.POST("/apply-for-withdrawal", middleware.JWTMiddleware(), withdrawAPI.ApplyForWithdrawal)

		// 回调接口
		withdraw.POST("/callback/kb", withdrawAPI.KBCallback)
		withdraw.POST("/callback/tk", withdrawAPI.TKCallback)
		withdraw.POST("/callback/at", withdrawAPI.ATCallback)
		withdraw.POST("/callback/go", withdrawAPI.GOCallback)
		withdraw.POST("/callback/cow", withdrawAPI.COWCallback)
		withdraw.POST("/callback/ant", withdrawAPI.ANTCallback)
		withdraw.POST("/callback/dy", withdrawAPI.DYCallback)
		withdraw.POST("/callback/gaga", withdrawAPI.GAGACallback)

		// 后台接口
		withdraw.POST("/admin/review-withdrawal", middleware.AdminMiddleware(), withdrawAPI.ReviewWithdrawal)
		withdraw.POST("/admin/add-user-withdraw-card", middleware.AdminMiddleware(), withdrawAPI.AddUserWithdrawCard)
		withdraw.POST("/admin/del-user-withdraw-card", middleware.AdminMiddleware(), withdrawAPI.DelUserWithdrawCard)
		withdraw.POST("/admin/fix-user-withdraw-card", middleware.AdminMiddleware(), withdrawAPI.FixUserWithdrawCard)
	}
}
