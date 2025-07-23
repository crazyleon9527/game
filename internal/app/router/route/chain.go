package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterChainRoutes(r *gin.RouterGroup, chainAPI *api.ChainAPI) {
	chain := r.Group("/blockchain")
	{
		chain.POST("/create-wallet-address", middleware.JWTMiddleware(), chainAPI.CreateWalletAddress)
		chain.POST("/get-wallet-address", middleware.JWTMiddleware(), chainAPI.GetWalletAddress)
		chain.POST("/get-wallet-address-list", middleware.JWTMiddleware(), chainAPI.GetWalletAddressList)
		chain.POST("/delete-wallet-address", middleware.JWTMiddleware(), chainAPI.DeleteWalletAddress)
		chain.POST("/get-blockchain-token", middleware.JWTMiddleware(), chainAPI.GetBlockchainToken)
		chain.POST("/get-blockchain-token-list", middleware.JWTMiddleware(), chainAPI.GetBlockchainTokenList)
	}
}
