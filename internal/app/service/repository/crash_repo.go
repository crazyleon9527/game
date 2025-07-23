package repository

import (
	"context"
	"errors"
	"fmt"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var CrashGameRepositorySet = wire.NewSet(wire.Struct(new(CrashGameRepository), "*"))

type CrashGameRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

// ------------------------------------ socket ------------------------------------

// 发布频道消息
func (r *CrashGameRepository) PublishChannelMessage(channel string, msg []byte) {
	r.RDS.Publish(context.Background(), fmt.Sprintf("crash_channel:%s", channel), msg)
}

// ------------------------------------ CrashGameRound ------------------------------------

func (r *CrashGameRepository) GetCrashGameRoundList() ([]*entities.CrashGameRound, error) {
	var list []*entities.CrashGameRound
	err := r.DB.Model(&entities.CrashGameRound{}).
		Where("settled = ?", constant.STATUS_SETTLE).
		Order("round_id desc").
		Limit(20).Find(&list).Error
	return list, err
}

func (r *CrashGameRepository) GetCrashGameRound(roundID uint64) (*entities.CrashGameRound, error) {
	var round entities.CrashGameRound
	err := r.DB.Model(&entities.CrashGameRound{}).
		Where("round_id = ?", roundID).
		First(&round).Error
	return &round, err
}

func (r *CrashGameRepository) GetLatestCrashGameRound() (*entities.CrashGameRound, error) {
	var round entities.CrashGameRound
	err := r.DB.Model(&entities.CrashGameRound{}).
		Order("round_id desc").
		First(&round).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &round, err
}

func (r *CrashGameRepository) CreateCrashGameRound(round *entities.CrashGameRound) error {
	return r.DB.Create(round).Error
}

func (r *CrashGameRepository) UpdateCrashGameRound(round *entities.CrashGameRound) error {
	return r.DB.Model(&entities.CrashGameRound{}).
		Where("round_id = ?", round.RoundID).
		Updates(round).Error
}

func (r *CrashGameRepository) UpdateSDGameRoundStatus(roundID string, status string) error {
	return r.DB.Model(&entities.CrashGameRound{}).
		Where("round_id = ?", roundID).
		Update("status", status).Error
}

// ------------------------------------ CrashGameOrder ------------------------------------

func (r *CrashGameRepository) GetCrashGameOrders(roundIDs []uint64) ([]*entities.CrashGameOrder, error) {
	var orders []*entities.CrashGameOrder
	err := r.DB.Model(&entities.CrashGameOrder{}).
		Where("round_id in ? and status = ?", roundIDs, constant.STATUS_CREATE).
		Find(&orders).Error
	return orders, err
}

func (r *CrashGameRepository) GetTopHeightCrashGameOrder(roundID uint64) (*entities.CrashGameOrder, error) {
	var order entities.CrashGameOrder
	err := r.DB.Model(&entities.CrashGameOrder{}).
		Where("round_id = ?", roundID).
		Order("escape_height desc").
		First(&order).Error
	return &order, err
}

func (r *CrashGameRepository) GetTopHeightCrashGameOrderList(roundIDs []uint64) ([]*entities.CrashGameOrder, error) {
	var orders []*entities.CrashGameOrder
	sql := `SELECT g1.*	FROM crash_game_order g1 INNER JOIN (SELECT round_id, MAX(escape_height) AS max_escape_height
		FROM crash_game_order WHERE round_id IN (?) GROUP BY round_id) g2 ON g1.round_id = g2.round_id AND g1.escape_height = g2.max_escape_height;`
	err := r.DB.Raw(sql, roundIDs).Scan(&orders).Error
	return orders, err
}

func (r *CrashGameRepository) GetUserCrashGameOrder(uid uint, roundID uint64) ([]*entities.CrashGameOrder, error) {
	var orders []*entities.CrashGameOrder
	err := r.DB.Model(&entities.CrashGameOrder{}).
		Where("uid = ? and round_id = ? and status = ?", uid, roundID, constant.STATUS_SETTLE).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *CrashGameRepository) GetUserCrashGameOrderList(uid uint) ([]*entities.CrashGameOrder, error) {
	var orders []*entities.CrashGameOrder
	err := r.DB.Model(&entities.CrashGameOrder{}).
		Where("uid = ? and status in (?)", uid, []int{constant.STATUS_CREATE, constant.STATUS_SETTLE}).
		Order("round_id desc").
		Limit(20).Find(&orders).Error
	return orders, err
}

func (r *CrashGameRepository) CreateCrashGameOrderWithTx(tx *gorm.DB, order *entities.CrashGameOrder) error {
	nowUnix := time.Now().Unix()
	isql := fmt.Sprintf(`INSERT INTO crash_game_order (created_at,updated_at,uid,name,round_id,bet_index,
		auto_escape_height,rate,bet_time,bet_amount,delivery,fee,pc,status) VALUES ('%d', '%d', '%d', '%s',
		 '%d', '%d', '%f', '%d', '%d', '%f', '%f', '%f', '%d', '%d') ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at),
		 auto_escape_height = VALUES(auto_escape_height), rate = VALUES(rate), bet_time = VALUES(bet_time), bet_amount = 
		 VALUES(bet_amount), delivery = VALUES(delivery), fee = VALUES(fee), pc = VALUES(pc), status = VALUES(status)`,
		nowUnix, nowUnix, order.UID, order.Name, order.RoundID, order.BetIndex, order.AutoEscapeHeight, order.Rate, order.BetTime,
		order.BetAmount, order.Delivery, order.Fee, order.PromoterCode, order.Status)
	return r.DB.Exec(isql).Error
}

func (r *CrashGameRepository) UpdateCrashGameOrderWithTx(tx *gorm.DB, order *entities.CrashGameOrder) error {
	return tx.Model(&entities.CrashGameOrder{}).
		Where("uid = ? and round_id = ? and bet_index = ?", order.UID, order.RoundID, order.BetIndex).
		Updates(order).Error
}

// ------------------------------------ CrashAutoBet ------------------------------------

func (r *CrashGameRepository) GetCrashAutoBetList(status int) ([]*entities.CrashAutoBet, error) {
	var autoBets []*entities.CrashAutoBet
	err := r.DB.Model(&entities.CrashAutoBet{}).
		Where("status = ?", status).
		Find(&autoBets).Error
	return autoBets, err
}

func (r *CrashGameRepository) GetCrashAutoBet(uid uint) (*entities.CrashAutoBet, error) {
	var autoBet entities.CrashAutoBet
	err := r.DB.Model(&entities.CrashAutoBet{}).
		Where("uid = ?", uid).
		First(&autoBet).Error
	return &autoBet, err
}

func (r *CrashGameRepository) CreateCrashAutoBet(autoBet *entities.CrashAutoBet) error {
	nowUnix := time.Now().Unix()
	isql := fmt.Sprintf(
		`INSERT INTO crash_auto_bet (created_at,updated_at,uid,bet_amount,auto_escape_height,auto_bet_count,status)
		VALUES ('%d', '%d', '%d', '%f', '%f', '%d', '%d') ON DUPLICATE KEY UPDATE bet_amount = VALUES(bet_amount),
		auto_escape_height = VALUES(auto_escape_height), auto_bet_count = VALUES(auto_bet_count), 
		status = VALUES(status), `, nowUnix, nowUnix, autoBet.UID, autoBet.BetAmount, autoBet.AutoEscapeHeight,
		autoBet.AutoBetCount, autoBet.Status)
	return r.DB.Exec(isql).Error
}

func (r *CrashGameRepository) UpdateCrashAutoBetStatus(uid uint, status uint8) error {
	return r.DB.Model(&entities.CrashAutoBet{}).
		Where("uid = ?", uid).
		Update("status", status).Error
}
