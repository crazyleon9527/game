package repository

import (
	"errors"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var DiceGameRepositorySet = wire.NewSet(wire.Struct(new(DiceGameRepository), "*"))

type DiceGameRepository struct {
	DB *gorm.DB
}

// ------------------------------------ DiceGameOrder ------------------------------------

func (r *DiceGameRepository) GetUserDiceGameOrder(uid uint) (*entities.DiceGameOrder, error) {
	var order entities.DiceGameOrder
	err := r.DB.Model(&entities.DiceGameOrder{}).
		Where("uid = ?", uid).
		Order("round_id desc").
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, err
}

func (r *DiceGameRepository) GetUserDiceGameOrderList(uid uint) ([]*entities.DiceGameOrder, error) {
	var orders []*entities.DiceGameOrder
	err := r.DB.Model(&entities.DiceGameOrder{}).
		Where("uid = ? and settled = ?", uid, constant.STATUS_SETTLE).
		Order("round_id desc").
		Limit(20).Find(&orders).Error
	return orders, err
}

func (r *DiceGameRepository) CreateDiceGameOrder(order *entities.DiceGameOrder) error {
	return r.CreateDiceGameOrderWithTx(r.DB, order)
}

func (r *DiceGameRepository) CreateDiceGameOrderWithTx(tx *gorm.DB, order *entities.DiceGameOrder) error {
	return tx.Create(order).Error
}

func (r *DiceGameRepository) UpdateDiceGameOrder(order *entities.DiceGameOrder) error {
	return r.UpdateDiceGameOrderWithTx(r.DB, order)
}

func (r *DiceGameRepository) UpdateDiceGameOrderWithTx(tx *gorm.DB, order *entities.DiceGameOrder) error {
	return tx.Model(&entities.DiceGameOrder{}).
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
			"pc":            order.PromoterCode,
			"settled":       order.Settled,
			"end_time":      order.EndTime,
		}).Error
}
