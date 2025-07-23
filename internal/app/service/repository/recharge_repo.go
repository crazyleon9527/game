package repository

import (
	"errors"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var RechargeRepositorySet = wire.NewSet(wire.Struct(new(RechargeRepository), "*"))

type RechargeRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

// func (r *RechargeRepository) GetRechargeListByUID(UID uint) ([]*entities.RechargeCard, error) {
// 	list := make([]*entities.RechargeCard, 0)
// 	err := r.DB.Where("uid = ? and status = 1", UID).Limit(10).Order("type desc").Find(&list).Error
// 	return list, err
// }

func (r *RechargeRepository) GetRechargeGoodList() ([]*entities.RechargeGood, error) {
	list := make([]*entities.RechargeGood, 0)
	err := r.DB.Where("status = ?", 1).Find(&list).Error
	return list, err
}

func (r *RechargeRepository) GetRechargeOrderList(param *entities.GetRechargeOrderListReq) error {
	var tx *gorm.DB = r.DB
	tx = tx.Where("uid = ? and status in(0,1,3,4)", param.UID).Order("id desc")
	param.List = make([]*entities.RechargeOrder, 0)
	return param.Paginate(tx)
}

func (r *RechargeRepository) GetRechargeSettingList() ([]*entities.RechargeSetting, error) {
	list := make([]*entities.RechargeSetting, 0)
	err := r.DB.Where("status = ? and recharge_state = ?", 1, 1).Order("sort DESC").Find(&list).Error
	return list, err
}

func (r *RechargeRepository) GetAvaliableRechargeSettingList() ([]*entities.RechargeSetting, error) {
	list := make([]*entities.RechargeSetting, 0)
	err := r.DB.Where("status = ?", 1).Order("sort DESC").Find(&list).Error
	return list, err
}

func (r *RechargeRepository) GetRechargeChannelSetting(entity *entities.RechargeChannelSetting) (*entities.RechargeChannelSetting, error) {
	result := r.DB.Last(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}

func (r *RechargeRepository) GetRechargeSetting(entity *entities.RechargeSetting) (*entities.RechargeSetting, error) {
	result := r.DB.Last(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}

func (r *RechargeRepository) UpdateRechargeSetting(entity *entities.RechargeSetting) error {
	return r.DB.Updates(entity).Error
}

func (r *RechargeRepository) GetRechargeActivity(entity *entities.RechargeActivity) (*entities.RechargeActivity, error) {
	result := r.DB.First(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}

func (r *RechargeRepository) CreateRechargeActivity(entity *entities.RechargeActivity) error {
	return r.DB.Create(entity).Error
}

func (r *RechargeRepository) UpdateRechargeActivity(entity *entities.RechargeActivity) error {
	return r.DB.Updates(entity).Error
}

func (r *RechargeRepository) GetRechargeOrder(entity *entities.RechargeOrder) (*entities.RechargeOrder, error) {
	// result := r.DB.Last(&entity, entity)

	result := r.DB.Clauses(dbresolver.Write).Last(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}

func (r *RechargeRepository) CreateRechargeOrder(entity *entities.RechargeOrder) error {
	return r.DB.Create(entity).Error
}

func (r *RechargeRepository) UpdateRechargeOrder(entity *entities.RechargeOrder) error {
	return r.DB.Updates(entity).Error
}

func (r *RechargeRepository) UpdateRechargeOrderWithTx(tx *gorm.DB, entity *entities.RechargeOrder) error {
	return tx.Updates(entity).Error
}

func (r *RechargeRepository) CreateCompletedRecharge(entity *entities.CompletedRecharge) error {
	return r.DB.Create(entity).Error
}

func (r *RechargeRepository) CreateCompletedRechargeWithTx(tx *gorm.DB, entity *entities.CompletedRecharge) error {
	return tx.Create(entity).Error
}

func (r *RechargeRepository) GetMinRecharge() (float64, error) {
	var gm entities.GmList
	result := r.DB.Select("min_recharge").First(&gm) // 1 是要查询的 ID
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, result.Error
	}
	return gm.MinRecharge, result.Error
}
