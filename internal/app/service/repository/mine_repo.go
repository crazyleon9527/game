package repository

import (
	"errors"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var MineGameRepositorySet = wire.NewSet(wire.Struct(new(MineGameRepository), "*"))

type MineGameRepository struct {
	DB *gorm.DB
}

// ------------------------------------ MineGameOrder ------------------------------------

func (r *MineGameRepository) GetUserMineGameOrder(uid uint) (*entities.MineGameOrder, error) {
	var order entities.MineGameOrder
	err := r.DB.Model(&entities.MineGameOrder{}).
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

func (r *MineGameRepository) GetUserMineGameOrderList(uid uint) ([]*entities.MineGameOrder, error) {
	var orders []*entities.MineGameOrder
	err := r.DB.Model(&entities.MineGameOrder{}).
		Where("uid = ? and settled = ?", uid, constant.STATUS_SETTLE).
		Order("round_id desc").
		Limit(20).Find(&orders).Error
	return orders, err
}

func (r *MineGameRepository) CreateMineGameOrder(order *entities.MineGameOrder) error {
	return r.CreateMineGameOrderWithTx(r.DB, order)
}

func (r *MineGameRepository) CreateMineGameOrderWithTx(tx *gorm.DB, order *entities.MineGameOrder) error {
	return tx.Create(order).Error
}

func (r *MineGameRepository) UpdateMineGameOrderStatus(uid uint, roundID uint64, status int) error {
	return r.DB.Model(&entities.MineGameOrder{}).
		Where("uid = ? and round_id = ?", uid, roundID).
		Update("status", status).Error
}

func (r *MineGameRepository) UpdateMineGameOrder(order *entities.MineGameOrder) error {
	return r.UpdateMineGameOrderWithTx(r.DB, order)
}

func (r *MineGameRepository) UpdateMineGameOrderWithTx(tx *gorm.DB, order *entities.MineGameOrder) error {
	return tx.Model(&entities.MineGameOrder{}).
		Where("uid = ? and round_id = ?", order.UID, order.RoundID).
		Updates(map[string]interface{}{
			"status":        order.Status,
			"client_seed":   order.ClientSeed,
			"server_seed":   order.ServerSeed,
			"mine_count":    order.MineCount,
			"diamond_left":  order.DiamondLeft,
			"mine_position": order.MinePosition,
			"open_position": order.OpenPosition,
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
