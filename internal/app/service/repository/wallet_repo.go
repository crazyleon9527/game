package repository

import (
	"context"
	"errors"
	"fmt"
	"rk-api/internal/app/entities"
	"rk-api/pkg/logger"
	"rk-api/pkg/rds"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var WalletRepositorySet = wire.NewSet(wire.Struct(new(WalletRepository), "*"))

type WalletRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

// func (r *WalletRepository) GetWalletList(param *entities.GetWalletListReq) error {
// 	var tx *gorm.DB = r.DB
// 	if param.Category != "" {
// 		tx = tx.Where("category = ?", param.Category)
// 	}
// 	tx = tx.Where("is_deleted = ?", 0)
// 	param.List = make([]*entities.Wallet, 0)
// 	return param.Paginate(tx)
// }

func (r *WalletRepository) CreateFundFreeze(entity *entities.FundFreeze) error {
	return r.DB.Create(entity).Error
}

func (r *WalletRepository) GetFundFreeze(entity *entities.FundFreeze) (*entities.FundFreeze, error) {
	result := r.DB.First(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}

func (r *WalletRepository) UpdateFundFreeze(tx *gorm.DB, entity *entities.FundFreeze) error {
	if err := tx.Model(&entities.FundFreeze{}).Updates(entity).Error; err != nil {
		return err
	}
	return nil
}

func (r *WalletRepository) UpdateWalletTTL(ID uint, expireTime time.Duration) error {

	if err := r.RDS.Expire(context.Background(), fmt.Sprintf("user:wallet:%d", ID), expireTime).Err(); err != nil {
		return err
	}
	return nil
}

func (r *WalletRepository) CreateWallet(user *entities.UserWallet) error {
	if err := r.DB.Create(user).Error; err != nil {
		return err
	}
	key := fmt.Sprintf("user:wallet:%d", user.UID)
	r.RDS.Del(context.Background(), key) // 删除旧的缓存
	return nil
}
func (r *WalletRepository) ClearWalletCache(uid uint) error {
	key := fmt.Sprintf("user:wallet:%d", uid)
	return r.RDS.Del(context.Background(), key).Err() // 删除旧的缓存
}

func (r *WalletRepository) UpdateCashWithTx(tx *gorm.DB, wallet *entities.UserWallet) error {

	logger.ZInfo("update wallet", zap.Any("wallet", wallet))
	key := fmt.Sprintf("user:wallet:%d", wallet.UID)

	err := tx.Model(&entities.UserWallet{}).
		Where("uid = ?", wallet.UID).
		Updates(map[string]interface{}{
			"cash": wallet.Cash, // 仅更新 cash 字段
		}).Error

	if err != nil {
		return err
	}

	redisKeyValues := map[string]string{
		"cash": fmt.Sprintf("%f", wallet.Cash),
	}
	return r.RDS.HSet(context.Background(), key, redisKeyValues).Err()

}

func (r *WalletRepository) UpdateWallet(tx *gorm.DB, wallet *entities.UserWallet) error {

	logger.ZInfo("update wallet", zap.Any("wallet", wallet))
	key := fmt.Sprintf("user:wallet:%d", wallet.UID)

	err := tx.Model(&entities.UserWallet{}).
		Where("uid = ?", wallet.UID).
		Select("cash", "diamond", "card", "promoter_code", "password", "security_level").
		Updates(wallet).Error

	if err != nil {
		return err
	}
	return r.RDS.Del(context.Background(), key).Err() // 删除旧的缓存

}

func (r *WalletRepository) GetWallet(uid uint) (*entities.UserWallet, error) {
	key := fmt.Sprintf("user:wallet:%d", uid)
	var wallet entities.UserWallet
	if err := r.RDS.HGetAll(context.Background(), key).Scan(&wallet); err == nil && wallet.UID == uid {
		return &wallet, nil
	}
	// 降级查数据库
	if err := r.DB.Where("uid = ?", uid).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if keyValues, err := rds.StructToRedisHashOptimized(wallet); err != nil {
		return nil, err
	} else {
		r.RDS.HSet(context.Background(), key, keyValues)
	}
	return &wallet, nil
}
