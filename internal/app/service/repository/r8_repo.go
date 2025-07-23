package repository

import (
	"errors"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var R8RepositorySet = wire.NewSet(wire.Struct(new(R8Repository), "*"))

type R8Repository struct {
	DB *gorm.DB
}

func (r *R8Repository) GetOrderByTransferNo(transferNo string, action string) (*entities.R8TransferOrder, error) {
	var entity entities.R8TransferOrder
	result := r.DB.Where("transfer_no = ? and action = ?", transferNo, action).Last(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *R8Repository) GetActivityOrderByAwardID(awardID string) (*entities.R8ActivityOrder, error) {
	var entity entities.R8ActivityOrder
	result := r.DB.Where("award_id = ?", awardID).Last(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *R8Repository) CreateOrder(entity *entities.R8TransferOrder) error {
	return r.DB.Create(entity).Error
}

func (r *R8Repository) CreateOrderWithTx(tx *gorm.DB, entity *entities.R8TransferOrder) error {
	return tx.Create(entity).Error
}

func (r *R8Repository) CreateActivityOrderWithTx(tx *gorm.DB, entity *entities.R8ActivityOrder) error {
	return tx.Create(entity).Error
}

func (r *R8Repository) GetLastestR8BetRecord() (*entities.R8BetRecord, error) {
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

func (r *R8Repository) BatchCreateR8BetRecord(list []*entities.R8BetRecord) error {
	return r.DB.CreateInBatches(list, len(list)).Error
}
