package repository

import (
	"fmt"
	"rk-api/internal/app/entities"
	"time"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var HashGameRepositorySet = wire.NewSet(wire.Struct(new(HashGameRepository), "*"))

type HashGameRepository struct {
	DB *gorm.DB
}

func (r *HashGameRepository) CreateHashGameOrderWithTx(tx *gorm.DB, order entities.IHashGameOrder) error {
	return tx.Create(order).Error
}

func (r *HashGameRepository) CreateHashGameRound(round entities.IHashGameRound) error {
	return r.DB.Create(round).Error
}

func (r *HashGameRepository) InitHashGameRound(round entities.IHashGameRound) error {
	nowUnix := time.Now().Unix()
	isql := fmt.Sprintf(`INSERT INTO sd_game_round (created_at,updated_at,roundId,blockHeight,status)
		VALUES ('%d', '%d', '%s', '%d', '%s') ON DUPLICATE KEY UPDATE status = VALUES(status)`,
		nowUnix, nowUnix, round.GetRoundID(), round.GetBlockHeight(), round.GetStatus())
	return r.DB.Exec(isql).Error
}

func (r *HashGameRepository) UpdateHashGameOrderWithTx(tx *gorm.DB, order entities.IHashGameOrder) error {
	return tx.Updates(order).Error
}

func (r *HashGameRepository) UpdateHashGameRound(round entities.IHashGameRound) error {
	return r.DB.Where("roundId = ?", round.GetRoundID()).Updates(round).Error
}

func (r *HashGameRepository) UpdateSDGameRoundStatus(roundID string, status string) error {
	return r.DB.Model(&entities.HashSDGameRound{}).Where("roundId = ?", roundID).Update("status", status).Error
}
