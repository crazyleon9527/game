package repository

import (
	"errors"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var ZfRepositorySet = wire.NewSet(wire.Struct(new(ZfRepository), "*"))

type ZfRepository struct {
	DB *gorm.DB
}

func (r *ZfRepository) GetOrderBy(uid uint, gameCode string, roundID int, betID int) (*entities.ZfTransferOrder, error) {
	var entity entities.ZfTransferOrder
	result := r.DB.Where("uid=? and game_code = ? and round_id = ? and bet_id = ?", uid, gameCode, roundID, betID).Last(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *ZfRepository) CreateOrder(entity *entities.ZfTransferOrder) error {
	return r.DB.Create(entity).Error
}

func (r *ZfRepository) CreateOrderWithTx(tx *gorm.DB, entity *entities.ZfTransferOrder) error {
	return tx.Create(entity).Error
}

func (r *ZfRepository) UpdateOrderWithTx(tx *gorm.DB, entity *entities.ZfTransferOrder) error {
	return tx.Updates(entity).Error
}

func (r *ZfRepository) GetLastestZfBetRecord() (*entities.ZfBetRecord, error) {
	var entity entities.ZfBetRecord
	result := r.DB.Last(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *ZfRepository) BatchCreateZfBetRecord(list []*entities.ZfBetRecord) error {
	return r.DB.CreateInBatches(list, len(list)).Error
}
