package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterWalletRoutes(r *gin.RouterGroup, walletAPI *api.WalletAPI) {
	wallet := r.Group("/wallet")
	{
		wallet.POST("/update-wallet-password", middleware.JWTMiddleware(), walletAPI.UpdateWalletPassword)
		wallet.POST("/enable-wallet-password", middleware.JWTMiddleware(), walletAPI.EnableWalletPassword)
		wallet.POST("/get-user-wallet", middleware.JWTMiddleware(), walletAPI.GetUserWallet)
	}
}
