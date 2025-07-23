package repository

import (
	"errors"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var FreeRepositorySet = wire.NewSet(wire.Struct(new(FreeRepository), "*"))

type FreeRepository struct {
	DB *gorm.DB
}

// func (r *FreeRepository) GetOrderByTransferNo(transferNo string, action string) (*entities.FreeTransferOrder, error) {
// 	var entity entities.FreeTransferOrder
// 	result := r.DB.Where("transfer_no = ? and action = ?", transferNo, action).Last(&entity)
// 	if result.Error != nil {
// 		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 			return nil, nil
// 		}
// 		return nil, result.Error
// 	}
// 	return &entity, nil
// }

func (r *FreeRepository) CreateFreeCard(entity *entities.FreeCard) error {
	return r.DB.Create(entity).Error
}

// 获取最低金额的可用免单卡
func (r *FreeRepository) GetAvailableFreeCard(uid uint, amount int64) (*entities.FreeCard, error) {
	var card entities.FreeCard
	if err := r.DB.Where("uid = ? and used = ? AND amount >= ?", uid, false, amount).
		Order("amount ASC").First(&card).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

// 更新免单卡
func (r *FreeRepository) Update(card *entities.FreeCard) error {
	return r.DB.Updates(card).Error
}

func (r *FreeRepository) GetFreeCard(id uint) (*entities.FreeCard, error) {
	var card entities.FreeCard
	// 降级查数据库
	if err := r.DB.First(&card, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &card, nil
}
