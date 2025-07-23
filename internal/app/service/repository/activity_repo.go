package repository

import (
	"errors"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var ActivityRepositorySet = wire.NewSet(wire.Struct(new(ActivityRepository), "*"))

type ActivityRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

func (r *ActivityRepository) GetActivityList() ([]*entities.Activity, error) {
	var activity []*entities.Activity
	// 查询数据库，获取所有区块链代币信息
	err := r.DB.Where("status = ? ", 1).Find(&activity).Error
	if err != nil {
		return nil, err
	}
	return activity, nil
}

func (r *ActivityRepository) GetBannerList() ([]*entities.Banner, error) {
	var banner []*entities.Banner
	// 查询数据库，获取所有banner
	err := r.DB.Where("status = ? ", 1).Order("priority desc").Find(&banner).Error
	if err != nil {
		return nil, err
	}
	return banner, nil
}

// GetLogoList
func (r *ActivityRepository) GetLogoList(logoType int) ([]*entities.Logo, error) {
	var logo []*entities.Logo
	// 查询数据库，获取所有banner
	err := r.DB.Where("status = ? and type = ?", 1, logoType).Find(&logo).Error
	if err != nil {
		return nil, err
	}
	return logo, nil
}

func (r *ActivityRepository) CreateHongbaoSetting(entity *entities.HongbaoSetting) error {
	return r.DB.Create(entity).Error
}

func (r *ActivityRepository) GetHongbaoSettingByName(name string) (*entities.HongbaoSetting, error) {
	entity := new(entities.HongbaoSetting)
	result := r.DB.Where("name = ? and status = 0", name).First(entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}

func (r *ActivityRepository) DelHongbaoSettingByName(name string) error {
	// return r.DB.Where("name = ? AND status = ?", name, 0).Delete(&entities.HongbaoSetting{}).Error

	return r.DB.Model(&entities.HongbaoSetting{}).Where("name = ? AND status = ?", name, 0).UpdateColumn("status", 1).Error
}

func (r *ActivityRepository) UpdatebaoSettingByWithTx(tx *gorm.DB, setting *entities.HongbaoSetting) error {
	return tx.Updates(setting).Error
}

func (r *ActivityRepository) GetHongbaoCountByHongID(hongID uint) (int64, error) {

	var count int64
	// 在读取红包数量时添加锁
	if err := r.DB.Clauses(dbresolver.Write).Set("gorm:query_option", "FOR UPDATE").Model(&entities.HongbaoRecord{}).Where("hong_id = ?", hongID).Count(&count).Error; err != nil {
		return 0, err
	}

	// if err := r.DB.Model(&entities.HongbaoRecord{}).Where("hong_id = ?", hongID).Count(&count).Error; err != nil {
	// 	return 0, err
	// }
	return count, nil
}

func (r *ActivityRepository) GetHongbao(entity *entities.HongbaoRecord) (*entities.HongbaoRecord, error) {
	result := r.DB.Last(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}
func (r *ActivityRepository) CreateHongbaoRecordWithTx(tx *gorm.DB, entity *entities.HongbaoRecord) error {
	return tx.Create(entity).Error
}

func (r *ActivityRepository) CreatePinduoRecord(pinduoRecord *entities.PinduoRecord) error {
	return r.DB.Create(pinduoRecord).Error
}

func (r *ActivityRepository) UpdatePinduoRecord(pinduoRecord *entities.PinduoRecord) error {
	return r.DB.Updates(pinduoRecord).Error
}

func (r *ActivityRepository) UpdatePinduoRecordWithTx(tx *gorm.DB, pinduoRecord *entities.PinduoRecord) error {
	return tx.Updates(pinduoRecord).Error
}

func (r *ActivityRepository) GetPinduoRecordByUID(uid uint) (*entities.PinduoRecord, error) {
	entity := new(entities.PinduoRecord)
	result := r.DB.Model(&entities.PinduoRecord{}).Where("uid = ?", uid).Last(entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
	}
	return entity, nil
}

func (r *ActivityRepository) GetPinduoSetting(entity *entities.PinduoSetting) (*entities.PinduoSetting, error) {
	result := r.DB.First(&entity, entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return entity, nil
}
