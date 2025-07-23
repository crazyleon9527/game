package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var NineRepositorySet = wire.NewSet(wire.Struct(new(NineRepository), "*"))

type NineRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

// func (r *NineRepository) GetRelationList(ID uint) ([]*entities.HallInviteRelation, error) {
// 	var tx *gorm.DB = r.DB
// 	tx = tx.Where("uid = ?", ID).Limit(9)
// 	list := make([]*entities.HallInviteRelation, 0)
// 	err := tx.Select("pid", "level").Offset(0).Limit(9).Find(&list).Error
// 	return list, err
// }

// func (r *NineRepository) AddRelationList(list []*entities.HallInviteRelation) error {
// 	return r.DB.CreateInBatches(list, len(list)).Error
// }

func (r *NineRepository) CreateNinePeriod(period *entities.NinePeriod) error {
	return r.DB.Create(period).Error
}

func (r *NineRepository) CreateNineOrderWithTx(tx *gorm.DB, order *entities.NineOrder) error {
	return tx.Create(order).Error
}

func (r *NineRepository) UpdateNineOrderWithTx(tx *gorm.DB, order *entities.NineOrder) error {
	return tx.Updates(order).Error
}

func (r *NineRepository) GetUnSettleNinePeriodList() ([]*entities.NinePeriod, error) {
	list := make([]*entities.NinePeriod, 0)
	err := r.DB.Where("status = ?", 0).Find(&list).Error
	// err := r.DB.Where("number = ?", -1).Find(&list).Error
	return list, err
}

// 获取未结算的nine订单
func (r *NineRepository) GetUnSettleNineOrderListByPeriodID(periodID string, betType uint8) ([]*entities.NineOrder, error) {
	list := make([]*entities.NineOrder, 0)
	err := r.DB.Clauses(dbresolver.Write).Where("period = ? and bet_type = ? and status = ?", periodID, betType, 0).Find(&list).Error
	return list, err
}

// 获取未结算的wingo订单
func (r *NineRepository) GetUnSettleExpiredNineOrder() (*entities.NineOrder, error) {

	var nineOrder entities.NineOrder

	currentTime := time.Now()
	fiveMinutesLater := currentTime.Add(-5 * time.Minute) //结算已经超过3分钟了。过期了
	fiveMinutesLaterUnix := fiveMinutesLater.Unix()

	result := r.DB.Where("status = ? AND finish_time < ?", 0, fiveMinutesLaterUnix).First(&nineOrder)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &nineOrder, nil
}

// 获取未结算的nine期数
func (r *NineRepository) GetNinePeriodByPeriodID(periodID string, betType uint8) (*entities.NinePeriod, error) {
	var period entities.NinePeriod
	result := r.DB.Where("period = ? and bet_type = ?", periodID, betType).Last(&period)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &period, nil
}

func (r *NineRepository) GetLastestNinePeriodByDate(periodDate string, betType uint8) (*entities.NinePeriod, error) {

	var period entities.NinePeriod
	result := r.DB.Where("period_date = ? and bet_type = ?", periodDate, betType).Last(&period)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &period, nil
}

func CreateNinePresetRDSKey(periodDate string) string {
	return fmt.Sprintf(constant.REDIS_NINE_PRESET, periodDate)
}

func (r *NineRepository) GetPresetNumberListRDS(periodDate string, betType string) ([]int, error) {
	res, err := r.RDS.HGet(context.Background(), CreateNinePresetRDSKey(periodDate), betType).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var list []int
	err = json.Unmarshal([]byte(res), &list)
	return list, err
}

func (r *NineRepository) AddPresetNumberListRDS(periodDate string, betType string, list []int) error {
	jsonArr, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = r.RDS.HSet(context.Background(), CreateNinePresetRDSKey(periodDate), betType, jsonArr).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *NineRepository) GetNineSettingList() ([]*entities.NineRoomSetting, error) {
	list := make([]*entities.NineRoomSetting, 0)
	err := r.DB.Where("status = ?", 1).Find(&list).Error
	return list, err
}

func (r *NineRepository) GetNineSetting(betType uint) (*entities.NineRoomSetting, error) {
	var setting entities.NineRoomSetting
	result := r.DB.Where("bet_type = ? and status = ?", betType, 1).Last(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &setting, nil
}

func (r *NineRepository) GetLastestNineOrderListByUID(param *entities.NineOrderHistoryReq) error {
	tx := r.DB.Where("uid = ? and bet_type = ?", param.UID, param.BetType).Order("created_at desc")
	param.List = make([]*entities.NineOrder, 0)
	return param.Paginate(tx)
}

func (r *NineRepository) GetLastestNinePeriodList(param *entities.GetPeriodHistoryListReq) error {
	tx := r.DB.Where("bet_type = ? and status = ?", param.BetType, 1).Order("created_at desc")
	param.List = make([]*entities.NinePeriod, 0)
	return param.Paginate(tx)
}

func (r *NineRepository) GetLastestPeriodHistoryListWithLimit(betType uint8, limit int) ([]*entities.NinePeriod, error) {
	list := make([]*entities.NinePeriod, 0)
	err := r.DB.Where("bet_type = ? and status = ?", betType, 1).Limit(limit).Order("created_at  DESC").Find(&list).Error
	return list, err
}

func (r *NineRepository) UpdateNinePeriod(period *entities.NinePeriod) error {
	err := r.DB.Updates(period).Error
	if err != nil {
		return err
	}
	if period.Number == 0 { //兼容为0的情况
		return r.DB.Model(period).UpdateColumn("number", period.Number).Error
	}
	return nil
}

func (r *NineRepository) GetTodayPeriodList(betType uint) ([]*entities.NinePeriod, error) {
	list := make([]*entities.NinePeriod, 0)
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	err := r.DB.Where("bet_type = ?", betType).
		Where("created_at >= ? ", startTime).
		// Order("created_at desc").
		Find(&list).Error
	return list, err
}

func (r *NineRepository) GetTodayPeriodTrend(betType uint) ([]*entities.NinePeriod, error) {
	list := make([]*entities.NinePeriod, 0)
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	err := r.DB.Where("bet_type = ? and status = ?", betType, 1).
		Where("created_at >= ? ", startTime).
		Select("number", "period_index").
		Order("created_at desc").
		Find(&list).Error

	return list, err
}
