package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var ChainAPISet = wire.NewSet(wire.Struct(new(ChainAPI), "*"))

type ChainAPI struct {
	Srv *service.ChainService
}

// 创建钱包地址（POST 请求）
func (c *ChainAPI) CreateWalletAddress(ctx *gin.Context) {
	var req entities.CreateWalletAddressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	walletAddress := entities.WalletAddress{
		UID:            ginx.Mine(ctx),
		BlockchainType: req.BlockchainType,
		TokenSymbol:    req.TokenSymbol,
		Address:        req.Address,
	}
	err := c.Srv.CreateWalletAddress(&walletAddress)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

// 获取钱包地址（POST 请求）
func (c *ChainAPI) GetWalletAddress(ctx *gin.Context) {
	var req entities.GetWalletAddressReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	req.UID = ginx.Mine(ctx)

	address, err := c.Srv.GetWalletAddress(req.UID, req.TokenSymbol, req.ChainType)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, address)
}

// 获取区块链代币列表（POST 请求）
func (c *ChainAPI) GetWalletAddressList(ctx *gin.Context) {
	tokens, err := c.Srv.GetWalletAddressList(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, tokens)
}

// 删除钱包地址（POST 请求）
func (c *ChainAPI) DeleteWalletAddress(ctx *gin.Context) {
	var req entities.IDReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Srv.DeleteWalletAddress(req.ID)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

// 获取区块链代币（POST 请求）
func (c *ChainAPI) GetBlockchainToken(ctx *gin.Context) {

	var req entities.GetBlockchainTokenReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	token, err := c.Srv.GetBlockchainToken(req.TokenSymbol, req.ChainType)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, token)
}

// 获取区块链代币列表（POST 请求）
func (c *ChainAPI) GetBlockchainTokenList(ctx *gin.Context) {
	tokens, err := c.Srv.GetBlockchainTokenList()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, tokens)
}
