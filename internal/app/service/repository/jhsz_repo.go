package repository

import (
	"errors"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var JhszRepositorySet = wire.NewSet(wire.Struct(new(JhszRepository), "*"))

type JhszRepository struct {
	DB *gorm.DB
}

func (r *JhszRepository) GetOrderByTransferNo(transferNo string, action string) (*entities.JhszTransferOrder, error) {
	var entity entities.JhszTransferOrder
	result := r.DB.Where("transfer_no = ? and action = ?", transferNo, action).Last(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *JhszRepository) CreateOrder(entity *entities.JhszTransferOrder) error {
	return r.DB.Create(entity).Error
}

func (r *JhszRepository) CreateOrderWithTx(tx *gorm.DB, entity *entities.JhszTransferOrder) error {
	return tx.Create(entity).Error
}

func (r *JhszRepository) UpdateOrderWithTx(tx *gorm.DB, entity *entities.JhszTransferOrder) error {
	return tx.Updates(entity).Error
}

func (r *JhszRepository) GetLastestR8BetRecord() (*entities.R8BetRecord, error) {
	var entity entities.R8BetRecord
	result := r.DB.Last(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *JhszRepository) BatchCreateR8BetRecord(list []*entities.R8BetRecord) error {
	return r.DB.CreateInBatches(list, len(list)).Error
}
