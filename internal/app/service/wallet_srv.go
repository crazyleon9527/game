package service

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/service/repository"
	"time"

	"github.com/google/wire"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var WalletServiceSet = wire.NewSet(
	ProvideWalletService,
)

type WalletService struct {
	Repo      *repository.WalletRepository
	UserLocks *entities.RedisUserLock
}

func ProvideWalletService(
	repo *repository.WalletRepository,

) *WalletService {
	service := &WalletService{
		Repo: repo,

		UserLocks: entities.NewRedisUserLock(repo.RDS), //分布式锁
	}
	return service
}

func (s *WalletService) UpdateWalletTTL(uid uint, expireTime time.Duration) error {
	return s.Repo.UpdateWalletTTL(uid, expireTime)
}

func (s *WalletService) GetWallet(uid uint) (*entities.UserWallet, error) {
	wallet, err := s.Repo.GetWallet(uid)

	return wallet, err
}

func (s *WalletService) CreateWallet(wallet *entities.UserWallet) error {
	return s.Repo.CreateWallet(wallet)
}

func (s *WalletService) UpdateWallet(wallet *entities.UserWallet) error {
	return s.Repo.UpdateWallet(s.Repo.DB, wallet)
}

func (s *WalletService) UpdateCashWithTx(tx *gorm.DB, wallet *entities.UserWallet) error {
	return s.Repo.UpdateCashWithTx(tx, wallet)
}

func (s *WalletService) UpdateWalletWithTx(tx *gorm.DB, wallet *entities.UserWallet) error {
	return s.Repo.UpdateWallet(tx, wallet)
}

func (s *WalletService) UpdateWalletCashWithTx(tx *gorm.DB, uid uint, cash float64) error {
	return s.Repo.UpdateWallet(tx, &entities.UserWallet{UID: uid, Cash: cash})
}

func (s *WalletService) GetUserWallet(uid uint) (*entities.UserWallet, error) {
	return s.GetWallet(uid)
}

func (s *WalletService) UpdateWalletPassword(uid uint, password, newPassword string) error {
	var wallet, err = s.GetWallet(uid)
	if err != nil {
		return err
	}
	if wallet == nil {
		return errors.WithCode(errors.WalletNotExist)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(wallet.Password), []byte(password)); err != nil {
		return errors.WithCode(errors.InvalidPassword)
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	wallet.Password = string(hashedNewPassword)
	wallet.SecurityLevel = 1
	if err := s.UpdateWallet(wallet); err != nil {
		return err
	}

	SendNotification(wallet.UID, "your wallet password has been changed, please remember your new password, do not disclose.", "Wallet Password Changed")
	return nil
}

func (s *WalletService) EnableWalletPassword(uid uint, password string) error {
	var wallet, err = s.GetWallet(uid)
	if err != nil {
		return err
	}
	if wallet == nil {
		return errors.WithCode(errors.WalletNotExist)
	}
	if wallet.Password != "" {
		return errors.WithCode(errors.WalletPasswordExist)
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	wallet.Password = string(hashedNewPassword)
	wallet.SecurityLevel = 1
	if err := s.UpdateWallet(wallet); err != nil {
		return err
	}

	SendNotification(wallet.UID, "your wallet password has been enabled, please remember your password, do not disclose.", "Wallet Password Enabled")
	return nil
}

// 举报冻结
func (s *WalletService) CreateFundFreeze(req *entities.FundFreeze) error {
	//校验参数
	return s.Repo.CreateFundFreeze(req)
}

func (s *WalletService) UpdateFundFreezeWithTx(tx *gorm.DB, entity *entities.FundFreeze) error {
	return s.Repo.UpdateFundFreeze(tx, entity)
}

func (s *WalletService) GetFundFreeze(req *entities.FundFreeze) (*entities.FundFreeze, error) {
	return s.Repo.GetFundFreeze(req)
}

/**
 * 钱包事务原子操作分布式锁
 * @param uid 用户ID
 * @param operation 业务操作
 * @return error
 */
func (s *WalletService) HandleWallet(uid uint, operation func(wallet *entities.UserWallet, tx *gorm.DB) error) (err error) {

	wallet, err := s.GetUserWallet(uid)
	if wallet == nil {
		return
	}
	// 开始事务
	tx := s.Repo.DB.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	// 加锁，防止并发访问
	s.UserLocks.Lock(uid)
	defer s.UserLocks.Unlock(uid)

	// 执行业务操作
	if err := operation(wallet, tx); err != nil {
		tx.Rollback() // 如果发生错误，回滚事务
		return err
	}

	if err = tx.Commit().Error; err != nil {
		return err
	}
	// 提交事务
	return nil
}

func (s *WalletService) ClearWalletCache(uid uint) error {
	return s.Repo.ClearWalletCache(uid)
}
