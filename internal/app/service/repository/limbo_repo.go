package repository

import (
	"errors"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var LimboGameRepositorySet = wire.NewSet(wire.Struct(new(LimboGameRepository), "*"))

type LimboGameRepository struct {
	DB *gorm.DB
}

// GetUserLimboGameOrder retrieves the latest Limbo game order for a user.
func (r *LimboGameRepository) GetUserLimboGameOrder(uid uint) (*entities.LimboGameOrder, error) {
	var order entities.LimboGameOrder
	err := r.DB.Model(&entities.LimboGameOrder{}).
		Where("uid = ?", uid).
		Order("round_id desc").
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetUserLimboGameOrderList retrieves a list of Limbo game orders for a user.
func (r *LimboGameRepository) GetUserLimboGameOrderList(uid uint) ([]*entities.LimboGameOrder, error) {
	var orders []*entities.LimboGameOrder
	err := r.DB.Model(&entities.LimboGameOrder{}).
		Where("uid = ?", uid).
		Order("round_id desc").
		Limit(20).Find(&orders).Error
	return orders, err
}

// CreateLimboGameOrder creates a new Limbo game order.
func (r *LimboGameRepository) CreateLimboGameOrder(order *entities.LimboGameOrder) error {
	return r.CreateLimboGameOrderWithTx(r.DB, order)
}

// CreateLimboGameOrderWithTx creates a new Limbo game order within a transaction.
func (r *LimboGameRepository) CreateLimboGameOrderWithTx(tx *gorm.DB, order *entities.LimboGameOrder) error {
	return tx.Create(order).Error
}

// UpdateLimboGameOrder updates an existing Limbo game order.
func (r *LimboGameRepository) UpdateLimboGameOrder(order *entities.LimboGameOrder) error {
	return r.UpdateLimboGameOrderWithTx(r.DB, order)
}

// UpdateLimboGameOrderWithTx updates an existing Limbo game order within a transaction.
func (r *LimboGameRepository) UpdateLimboGameOrderWithTx(tx *gorm.DB, order *entities.LimboGameOrder) error {
	return tx.Model(&entities.LimboGameOrder{}).
		Where("uid = ? and round_id = ?", order.UID, order.RoundID).
		Updates(map[string]interface{}{
			"client_seed":   order.ClientSeed,
			"server_seed":   order.ServerSeed,
			"target":        order.Target,
			"result":        order.Result,
			"is_above":      order.IsAbove,
			"multiple":      order.Multiple,
			"rate":          order.Rate,
			"bet_time":      order.BetTime,
			"bet_amount":    order.BetAmount,
			"delivery":      order.Delivery,
			"fee":           order.Fee,
			"reward_amount": order.RewardAmount,
			"settled":       order.Settled,
			"end_time":      order.EndTime,
		}).Error
}