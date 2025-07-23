package repository

import (
	"context"
	"fmt"
	"rk-api/internal/app/entities"
	"strconv"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	gameOnlineCacheKey = "game:online" // 单游戏在线人数缓存键
)

var GameRepositorySet = wire.NewSet(wire.Struct(new(GameRepository), "*"))

type GameRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}
type GameRefresh struct {
	GameCode    string `gorm:"size:64" json:"game_code"` // 游戏编码
	OnlineCount int    `json:"online_count"`
}

func (r *GameRepository) GetGameRefreshList() ([]*entities.GameRefresh, error) {
	results, err := r.RDS.HGetAll(context.Background(), "game:online").Result()
	if err != nil {
		return nil, err
	}

	var list []*entities.GameRefresh
	for gameCode, countStr := range results {
		count, err := strconv.Atoi(countStr)
		if err != nil {
			continue // 转换失败跳过
		}
		list = append(list, &entities.GameRefresh{
			GameCode:    gameCode,
			OnlineCount: count,
		})
	}
	return list, nil
}

func (r *GameRepository) GetOnlineCountFromCache(game string) (int, error) {
	count, err := r.RDS.HGet(context.Background(), gameOnlineCacheKey, game).Int()
	if err == redis.Nil {
		return 0, nil // 若无数据，则返回 0
	}
	return count, err
}

func (r *GameRepository) CacheGameRefreshList(list []*entities.GameRefresh, ttl time.Duration) error {
	ctx := context.Background()
	pipe := r.RDS.Pipeline()

	for _, game := range list {
		pipe.HSet(ctx, gameOnlineCacheKey, game.GameCode, game.OnlineCount)
	}

	// 如果需要全局过期时间（可选）
	pipe.Expire(ctx, gameOnlineCacheKey, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *GameRepository) SyncOnlineCount2DB(list []*entities.GameRefresh) error {
	// 使用事务保证原子性
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 批量更新语句构造（适用于MySQL）
		var caseStatements []string
		var params []interface{}

		for _, game := range list {
			caseStatements = append(caseStatements, "WHEN ? THEN ?")
			params = append(params, game.GameCode, game.OnlineCount)
		}

		query := fmt.Sprintf(`
			UPDATE games
			SET online_count = CASE game_code %s END
			WHERE game_code IN (?)
		`, strings.Join(caseStatements, " "))

		// 收集所有ID
		var gameCodes []string
		for _, game := range list {
			gameCodes = append(gameCodes, game.GameCode)
		}
		params = append(params, gameCodes)

		// 执行批量更新
		if err := tx.Exec(query, params...).Error; err != nil {
			return err
		}
		// 失效相关缓存
		return nil
	})
}

func (r *GameRepository) GetGameList(param *entities.GetGameListReq) error {
	var tx *gorm.DB = r.DB
	if param.Category != "" {
		tx = tx.Where("category = ?", param.Category)
	}
	tx = tx.Where("is_deleted = ? and is_active = ?", 0, 1).Order("priority desc")
	param.List = make([]*entities.Game, 0)
	return param.Paginate(tx)
}

func (r *GameRepository) GetWholeGameList() ([]*entities.Game, error) {
	var games []*entities.Game
	// 查询数据库，获取所有区块链代币信息
	err := r.DB.Find(&games).Error
	if err != nil {
		return nil, err
	}
	return games, nil
}

func (r *GameRepository) SearchGameList(param *entities.SearchGameReq) error {
	var tx *gorm.DB = r.DB
	if param.Category != "" {
		tx = tx.Where("category = ?", param.Category)
	}
	tx = tx.Where("is_deleted = ? and is_active = ?", 0, 1)
	if param.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+param.Name+"%")
	}
	param.List = make([]*entities.Game, 0)
	return param.Paginate(tx)
}
