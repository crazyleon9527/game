package repository

import (
	"context"
	"errors"
	"fmt"
	"rk-api/internal/app/entities"
	"rk-api/pkg/rds"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var FinancialRepositorySet = wire.NewSet(wire.Struct(new(FinancialRepository), "*"))

type FinancialRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

func (r *FinancialRepository) CreateFinancialSummary(user *entities.FinancialSummary) error {
	if err := r.DB.Create(user).Error; err != nil {
		return err
	}
	key := fmt.Sprintf("user:finance:%d", user.UID)
	r.RDS.Del(context.Background(), key) // 删除旧的缓存
	return nil
}

func (r *FinancialRepository) GetSummary(uid uint) (*entities.FinancialSummary, error) {

	var summary entities.FinancialSummary
	key := fmt.Sprintf("user:finance:%d", uid)
	if err := r.RDS.HGetAll(context.Background(), key).Scan(&summary); err == nil {
		return &summary, nil
	}

	// 降级查数据库

	if err := r.DB.First(&summary, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if keyValues, err := rds.StructToRedisHashOptimized(summary); err != nil {
		return nil, err
	} else {
		r.RDS.HSet(context.Background(), key, keyValues)
	}

	return &summary, nil
}
