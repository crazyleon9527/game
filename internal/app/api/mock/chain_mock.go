package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// ChainSet 注入Chain
var ChainSet = wire.NewSet(wire.Struct(new(Chain), "*"))

type Chain struct {
}

// @Summary 创建钱包地址
// @Description 创建新的钱包地址
// @Tags Chain
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param wallet_address body entities.WalletAddress true "钱包地址"
// @Router /api/blockchain/create-wallet-address [post]
func (c *Chain) CreateWalletAddress(ctx *gin.Context) {

}

// @Summary 获取钱包地址
// @Description 获取指定用户的区块链钱包地址
// @Tags Chain
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.GetWalletAddressReq true "查询条件"
// @Success 200 {object} entities.WalletAddress "钱包地址信息"
// @Router /api/blockchain/get-wallet-address [post]
func (c *Chain) GetWalletAddress(ctx *gin.Context) {

}

// @Summary 获取用户的钱包地址列表
// @Description 获取用户的钱包地址列表
// @Tags Chain
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} entities.WalletAddress "用户钱包地址"
// @Router /api/blockchain/get-wallet-address-list [post]
func (c *Chain) GetWalletAddressList(ctx *gin.Context) {

}

// @Summary 删除钱包地址
// @Description 删除指定钱包地址
// @Tags Chain
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.IDReq true "钱包地址ID"
// @Router /api/blockchain/delete-wallet-address [post]
func (c *Chain) DeleteWalletAddress(ctx *gin.Context) {

}

// @Summary 获取区块链代币
// @Description 获取指定区块链代币的信息
// @Tags Chain
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.GetBlockchainTokenReq true "查询条件"
// @Success 200 {object} entities.BlockchainToken "区块链代币信息"
// @Router /api/blockchain/get-blockchain-token [post]
func (c *Chain) GetBlockchainToken(ctx *gin.Context) {

}

// @Summary 获取区块链代币列表
// @Description 获取所有区块链代币列表
// @Tags Chain
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} entities.BlockchainToken "区块链代币信息"
// @Router /api/blockchain/get-blockchain-token-list [post]
func (c *Chain) GetBlockchainTokenList(ctx *gin.Context) {

}
