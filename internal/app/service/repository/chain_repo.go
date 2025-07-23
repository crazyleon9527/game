package repository

import (
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ChainRepositorySet = wire.NewSet(wire.Struct(new(ChainRepository), "*"))

type ChainRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

func (r *ChainRepository) Create(walletAddress *entities.WalletAddress) error {
	return r.DB.Create(walletAddress).Error
}

// GetByType 根据地址类型获取钱包地址
func (r *ChainRepository) GetWalletAddress(uid uint, tokenSymbol, chainType string) (*entities.WalletAddress, error) {
	var address entities.WalletAddress
	err := r.DB.Where("uid = ? AND token_symbol = ? AND blockchain_type = ?", uid, tokenSymbol, chainType).First(&address).Error
	return &address, err
}

func (r *ChainRepository) GetWalletAddressList(uid uint) ([]*entities.WalletAddress, error) {
	var WalletAddresss []*entities.WalletAddress
	// 查询数据库，获取所有区块链代币信息
	err := r.DB.Where("uid = ? ", uid).Find(&WalletAddresss).Error
	if err != nil {
		return nil, err
	}
	return WalletAddresss, nil
}

// Delete 删除钱包地址
func (r *ChainRepository) Delete(addressID uint) error {
	return r.DB.Delete(&entities.WalletAddress{}, addressID).Error
}

func (r *ChainRepository) CreateBlockchainToken(blockchainToken *entities.BlockchainToken) error {
	return r.DB.Create(blockchainToken).Error
}

func (r *ChainRepository) GetBlockchainToken(tokenSymbol, chainType string) (*entities.BlockchainToken, error) {
	var blockchainToken entities.BlockchainToken
	err := r.DB.Where("token_symbol = ? AND blockchain_type = ?", tokenSymbol, chainType).First(&blockchainToken).Error
	return &blockchainToken, err
}

func (r *ChainRepository) GetBlockchainTokenList() ([]*entities.BlockchainToken, error) {
	var blockchainTokens []*entities.BlockchainToken
	// 查询数据库，获取所有区块链代币信息
	err := r.DB.Where("active = ? ", 1).Find(&blockchainTokens).Error
	if err != nil {
		return nil, err
	}
	return blockchainTokens, nil
}
