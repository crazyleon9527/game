package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// WalletSet 注入Wallet
var WalletSet = wire.NewSet(wire.Struct(new(Wallet), "*"))

type Wallet struct {
}

// @Tags Wallet
// @Summary 更新资金密码
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.UpdateWalletPasswordReq true "params"
// @Success 200
// @Router /api/wallet/update-wallet-password [post]
func (a *Wallet) UpdateWalletPassword(c *gin.Context) {
}

// @Tags Wallet
// @Summary 启用资金密码
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.EnableWalletPasswordReq true "params"
// @Success 200
// @Router /api/wallet/enable-wallet-password [post]
func (a *Wallet) EnableWalletPassword(c *gin.Context) {
}

// @Tags Wallet
// @Summary 获取钱包
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200
// @Success 200 {object} entities.UserWallet
// @Router /api/wallet/get-user-wallet [post]
func (a *Wallet) GetUserWallet(c *gin.Context) {
}
