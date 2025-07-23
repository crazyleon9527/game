package service

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"

	"github.com/google/wire"
)

var ChainServiceSet = wire.NewSet(
	ProvideChainService,
)

// ChainService 管理应用程序的状态
type ChainService struct {
	Repo   *repository.ChainRepository
	locker *entities.RedisUserLock
}

// NewChainService 创建并返回一个新的 ChainService 实例
func ProvideChainService(repo *repository.ChainRepository) *ChainService {
	service := &ChainService{
		Repo:   repo,
		locker: entities.NewRedisUserLock(repo.RDS),
	}
	// Validate merchant codes on startup
	return service
}

// 创建个人钱包地址
func (s *ChainService) CreateWalletAddress(walletAddress *entities.WalletAddress) error {
	if err := utils.ValidateAddress(walletAddress.Address, walletAddress.BlockchainType); err != nil {
		return err
	}
	return s.Repo.Create(walletAddress)
}

// 获取钱包地址
func (s *ChainService) GetWalletAddress(uid uint, tokenSymbol, chainType string) (*entities.WalletAddress, error) {
	return s.Repo.GetWalletAddress(uid, tokenSymbol, chainType)
}

func (s *ChainService) GetWalletAddressList(uid uint) ([]*entities.WalletAddress, error) {
	return s.Repo.GetWalletAddressList(uid)
}

// 删除钱包地址
func (s *ChainService) DeleteWalletAddress(addressID uint) error {
	return s.Repo.Delete(addressID)
}

// 创建区块链代币
func (s *ChainService) CreateBlockchainToken(blockchainToken *entities.BlockchainToken) error {
	return s.Repo.CreateBlockchainToken(blockchainToken)
}

// 获取区块链代币
func (s *ChainService) GetBlockchainToken(tokenSymbol, chainType string) (*entities.BlockchainToken, error) {
	return s.Repo.GetBlockchainToken(tokenSymbol, chainType)
}

// 获取区块链代币列表
func (s *ChainService) GetBlockchainTokenList() ([]*entities.BlockchainToken, error) {
	return s.Repo.GetBlockchainTokenList()
}
