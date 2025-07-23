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

var WingoRepositorySet = wire.NewSet(wire.Struct(new(WingoRepository), "*"))

type WingoRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

// func (r *WingoRepository) GetRelationList(ID uint) ([]*entities.HallInviteRelation, error) {
// 	var tx *gorm.DB = r.DB
// 	tx = tx.Where("uid = ?", ID).Limit(9)
// 	list := make([]*entities.HallInviteRelation, 0)
// 	err := tx.Select("pid", "level").Offset(0).Limit(9).Find(&list).Error
// 	return list, err
// }

// func (r *WingoRepository) AddRelationList(list []*entities.HallInviteRelation) error {
// 	return r.DB.CreateInBatches(list, len(list)).Error
// }

func (r *WingoRepository) CreateWingoPeriod(period *entities.WingoPeriod) error {
	return r.DB.Create(period).Error
}

func (r *WingoRepository) CreateWingoOrderWithTx(tx *gorm.DB, order *entities.WingoOrder) error {
	return tx.Create(order).Error
}

func (r *WingoRepository) UpdateWingoOrderWithTx(tx *gorm.DB, order *entities.WingoOrder) error {
	return tx.Updates(order).Error
}

func (r *WingoRepository) GetUnSettleWingoPeriodList() ([]*entities.WingoPeriod, error) {
	list := make([]*entities.WingoPeriod, 0)
	err := r.DB.Where("status = ?", 0).Find(&list).Error
	// err := r.DB.Where("number = ?", -1).Find(&list).Error
	return list, err
}

// 获取未结算的wingo订单
func (r *WingoRepository) GetUnSettleWingoOrderListByPeriodID(periodID string, betType uint8) ([]*entities.WingoOrder, error) {
	list := make([]*entities.WingoOrder, 0)
	err := r.DB.Clauses(dbresolver.Write).Where("period = ? and bet_type = ? and status = ?", periodID, betType, 0).Find(&list).Error
	return list, err
}

// 获取未结算的wingo订单
func (r *WingoRepository) GetUnSettleExpiredWingoOrder() (*entities.WingoOrder, error) {

	var wingoOrder entities.WingoOrder

	currentTime := time.Now()
	fiveMinutesLater := currentTime.Add(-5 * time.Minute) //结算已经超过3分钟了。过期了
	fiveMinutesLaterUnix := fiveMinutesLater.Unix()

	result := r.DB.Where("status = ? AND finish_time < ?", 0, fiveMinutesLaterUnix).First(&wingoOrder)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &wingoOrder, nil
}

// 获取未结算的wingo期数
func (r *WingoRepository) GetWingoPeriodByPeriodID(periodID string, betType uint8) (*entities.WingoPeriod, error) {
	var period entities.WingoPeriod
	result := r.DB.Where("period = ? and bet_type = ?", periodID, betType).Last(&period)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &period, nil
}

func (r *WingoRepository) GetLastestWingoPeriodByDate(periodDate string, betType uint8) (*entities.WingoPeriod, error) {

	var period entities.WingoPeriod
	result := r.DB.Where("period_date = ? and bet_type = ?", periodDate, betType).Last(&period)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &period, nil
}

func CreateWingoPresetRDSKey(periodDate string) string {
	return fmt.Sprintf(constant.REDIS_WINGO_PRESET, periodDate)
}

func (r *WingoRepository) GetPresetNumberListRDS(periodDate string, betType string) ([]int, error) {
	res, err := r.RDS.HGet(context.Background(), CreateWingoPresetRDSKey(periodDate), betType).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var list []int
	err = json.Unmarshal([]byte(res), &list)
	return list, nil
}

func (r *WingoRepository) AddPresetNumberListRDS(periodDate string, betType string, list []int) error {
	jsonArr, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = r.RDS.HSet(context.Background(), CreateWingoPresetRDSKey(periodDate), betType, jsonArr).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *WingoRepository) GetWingoSettingList() ([]*entities.WingoRoomSetting, error) {
	list := make([]*entities.WingoRoomSetting, 0)
	err := r.DB.Where("status = ?", 1).Find(&list).Error
	return list, err
}

func (r *WingoRepository) GetWingoSetting(betType uint) (*entities.WingoRoomSetting, error) {
	var setting entities.WingoRoomSetting
	result := r.DB.Where("bet_type = ? and status = ?", betType, 1).Last(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &setting, nil
}

func (r *WingoRepository) GetLastestWingoOrderListByUID(param *entities.WingoOrderHistoryReq) error {
	tx := r.DB.Where("uid = ? and bet_type = ?", param.UID, param.BetType).Order("created_at desc")
	param.List = make([]*entities.WingoOrder, 0)
	return param.Paginate(tx)
}

func (r *WingoRepository) GetLastestWingoPeriodList(param *entities.GetPeriodHistoryListReq) error {
	tx := r.DB.Where("bet_type = ? and status = ?", param.BetType, 1).Order("created_at desc")
	param.List = make([]*entities.WingoPeriod, 0)
	return param.Paginate(tx)
}

func (r *WingoRepository) GetLastestPeriodHistoryListWithLimit(betType uint8, limit int) ([]*entities.WingoPeriod, error) {
	list := make([]*entities.WingoPeriod, 0)
	err := r.DB.Where("bet_type = ? and status = ?", betType, 1).Limit(limit).Order("created_at  DESC").Find(&list).Error
	return list, err
}

func (r *WingoRepository) UpdateWingoPeriod(period *entities.WingoPeriod) error {
	err := r.DB.Updates(period).Error
	if err != nil {
		return err
	}
	if period.Number == 0 { //兼容为0的情况
		return r.DB.Model(period).UpdateColumn("number", period.Number).Error
	}
	return nil
}

func (r *WingoRepository) GetTodayPeriodList(param *entities.GetPeriodListReq) error {
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	tx := r.DB.Where("bet_type = ?", param.BetType).Where("created_at >= ? ", startTime)
	param.List = make([]*entities.WingoPeriod, 0)
	return param.Paginate(tx)
}

func (r *WingoRepository) GetAllTodayPeriodList(betType uint) ([]*entities.WingoPeriod, error) {
	list := make([]*entities.WingoPeriod, 0)
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	err := r.DB.Where("bet_type = ?", betType).
		Where("created_at >= ? ", startTime).
		// Order("created_at desc").
		Find(&list).Error
	return list, err
}

func (r *WingoRepository) GetTodayPeriodTrend(betType uint) ([]*entities.WingoPeriod, error) {
	list := make([]*entities.WingoPeriod, 0)
	now := time.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	err := r.DB.Where("bet_type = ? and status = ?", betType, 1).
		Where("created_at >= ? ", startTime).
		Select("number", "period_index").
		Order("created_at desc").
		Find(&list).Error
	return list, err
}
