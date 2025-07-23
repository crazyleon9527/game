package repository

import (
	"errors"
	"rk-api/internal/app/entities"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var WithdrawRepositorySet = wire.NewSet(wire.Struct(new(WithdrawRepository), "*"))

type WithdrawRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

func (r *WithdrawRepository) GetWithdrawCard(entity *entities.WithdrawCard) (*entities.WithdrawCard, error) {
	result := r.DB.Last(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}
func (r *WithdrawRepository) CreateWithdrawCard(entity *entities.WithdrawCard) error {
	return r.DB.Create(entity).Error
}

func (r *WithdrawRepository) GetWithdrawCardByID(id uint) (*entities.WithdrawCard, error) {
	var entity entities.WithdrawCard
	result := r.DB.Where("id = ?", id).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *WithdrawRepository) UpdateWithdrawCard(entity *entities.WithdrawCard) error {
	return r.DB.Updates(entity).Error
}

func (r *WithdrawRepository) DelWithdrawCardByID(uid uint, id uint) error {
	// return r.DB.Model(&entities.WithdrawCard{}).Where("id = ? and uid = ?", id, uid).Update("status", 0).Error

	return r.DB.Where("id = ? and uid = ?", id, uid).Limit(1).Delete(&entities.WithdrawCard{}).Error
}

func (r *WithdrawRepository) SelectWithdrawCard(uid uint, id uint) error {
	// 找到uid为用户所选的uid，active为1（正在使用）的卡，将其设置为0（即，不使用）
	if err := r.DB.Model(&entities.WithdrawCard{}).Where("uid = ? AND active = ?", uid, 1).Update("active", 0).Error; err != nil {
		return err
	}
	// 然后将用户所选择的卡片（根据id）设置为正在使用（type为1）
	if err := r.DB.Model(&entities.WithdrawCard{}).Where("uid = ? AND id = ?", uid, id).Update("active", 1).Error; err != nil {
		return err
	}
	return nil
}

func (r *WithdrawRepository) GetWithdrawCardListByUID(UID uint) ([]*entities.WithdrawCard, error) {
	list := make([]*entities.WithdrawCard, 0)
	err := r.DB.Where("uid = ? and status = 1", UID).Limit(10).Order("active desc").Find(&list).Error
	return list, err
}

func (r *WithdrawRepository) CountWithdrawCardsByUID(UID uint) (int64, error) {
	var count int64
	err := r.DB.Model(&entities.WithdrawCard{}).Where("uid = ? and status = 1", UID).Count(&count).Error
	return count, err
}

func (r *WithdrawRepository) GetHallWithdrawRecordList(param *entities.GetHallWithdrawRecordListReq) error {
	var tx *gorm.DB = r.DB
	tx = tx.Where("uid = ? and status in(0,1,3,4)", param.UID).Order("id desc")
	param.List = make([]*entities.HallWithdrawRecord, 0)
	return param.Paginate(tx)
}

func (r *WithdrawRepository) GetTodayWithdrawCount(uid uint) (int64, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var count int64
	if err := r.DB.Model(&entities.HallWithdrawRecord{}).Where("uid = ? and status in(0,1,3,4)", uid).Where("created_at >= ?", today.Unix()).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *WithdrawRepository) GetHallWithdrawRecord(entity *entities.HallWithdrawRecord) (*entities.HallWithdrawRecord, error) {
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

func (r *WithdrawRepository) GetWithdrawCardRecordByID(id uint) (*entities.HallWithdrawRecord, error) {
	var entity entities.HallWithdrawRecord
	result := r.DB.Clauses(dbresolver.Write).Where("id = ?", id).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &entity, nil
}

func (r *WithdrawRepository) CreateHallWithdrawRecordWithTx(tx *gorm.DB, entity *entities.HallWithdrawRecord) error {
	return tx.Create(entity).Error
}

func (r *WithdrawRepository) UpdateHallWithdrawRecord(entity *entities.HallWithdrawRecord) error {
	return r.DB.Updates(entity).Error
}

func (r *WithdrawRepository) UpdateHallWithdrawRecordWithTx(tx *gorm.DB, entity *entities.HallWithdrawRecord) error {
	return tx.Updates(entity).Error
}

func (r *WithdrawRepository) GetRechargeSetting(entity *entities.RechargeSetting) (*entities.RechargeSetting, error) {
	result := r.DB.Last(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}

func (r *WithdrawRepository) GetRechargeChannelSetting(entity *entities.RechargeChannelSetting) (*entities.RechargeChannelSetting, error) {
	result := r.DB.Last(&entity, entity)
	if result.Error != nil {
		// if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 	return nil, nil
		// }
		return nil, result.Error
	}
	return entity, nil
}

func (r *WithdrawRepository) CreateCompletedWithdrawWithTx(tx *gorm.DB, entity *entities.CompletedWithdraw) error {
	return tx.Create(entity).Error
}

func (r *WithdrawRepository) GetMinWithdraw() (float64, error) {
	var gm entities.GmList
	result := r.DB.Select("min_withdraw").First(&gm) // 1 是要查询的 ID
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, result.Error
	}
	return gm.MinWithdraw, result.Error
}
