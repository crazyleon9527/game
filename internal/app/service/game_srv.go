package service

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
	"go.uber.org/zap"
)

const (
	QUERY_KEY_GAME_LIST = "QUERY_KEY_GAME_LIST"
)

type IFetchOnlineCount interface {
	// 修改接口定义，移除playerID参数
	FetchOnlineCount() ([]map[string]interface{}, error)
}

var GameServiceSet = wire.NewSet(
	ProvideGameService,
)

type GameService struct {
	Repo     *repository.GameRepository
	adapters map[string]IFetchOnlineCount
	UserSrv  *UserService

	GameCategoryCache *ecache.Cache // 新建的缓存，用来存储 gameCode 和 category

	cacheTTL time.Duration
}

// ProvideGameService 创建 GameService 实例
func ProvideGameService(
	repo *repository.GameRepository,
	userSrv *UserService,
	JhszSrv *JhszService,
) *GameService {
	adapters := make(map[string]IFetchOnlineCount)
	adapters["jhsz"] = JhszSrv
	service := &GameService{
		Repo:              repo,
		UserSrv:           userSrv,
		GameCategoryCache: ecache.NewLRUCache(3, 20, 60*time.Minute), // 初始化新缓存
	}
	return service
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// 同步所有第三方游戏数据
func (s *GameService) SyncThirdOnlineCount() {
	// _, cancel := context.WithTimeout(context.Background(), 110*time.Second)
	// defer cancel()
	if len(s.adapters) == 0 {
		return
	}

	for source, adapter := range s.adapters {

		s.SyncOnlineCount(source, adapter)
	}
}

// CacheGameIdentification 将 gameCode 和 Category 映射存入新缓存
func (s *GameService) CacheGameIdentification() error {
	defer utils.PrintPanicStack()
	games, err := s.Repo.GetWholeGameList()
	if err != nil {
		return err
	}
	for _, game := range games {
		logger.ZInfo("CacheGameCategory", zap.String("game_code", game.GameCode), zap.String("category", game.Category))
		s.GameCategoryCache.Put(game.GameCode, &entities.GameIdentification{
			Category: game.Category,
			Name:     game.Name,
		}) // 缓存10分钟
	}
	s.GameCategoryCache.Put("GameCategoryCache", 1) // 缓存10分钟
	return nil
}

// GetGameCategory 从新缓存中获取游戏的 Category
func (s *GameService) GetGameIdentification(gameCode string) *entities.GameIdentification {
	// 从缓存中获取 category
	if _, ok := s.GameCategoryCache.Get("GameCategoryCache"); !ok {
		s.CacheGameIdentification() // 缓存过期，重新缓存
	}

	cachedCategory, found := s.GameCategoryCache.Get(gameCode)
	if !found {
		return nil
	}
	// 将缓存的值转化为 string 类型
	category, ok := cachedCategory.(*entities.GameIdentification)
	if !ok {
		return nil
	}
	return category
}

func (s *GameService) GetGameList(req *entities.GetGameListReq) error {
	return s.Repo.GetGameList(req)
}

func (s *GameService) SearchGame(req *entities.SearchGameReq) error {
	return s.Repo.SearchGameList(req)
}

func (s *GameService) GetRefreshGameList() ([]*entities.GameRefresh, error) {
	// 尝试从 Redis 获取缓存
	return s.Repo.GetGameRefreshList()
}

// 真实同步第三方游戏数据
func (s *GameService) SyncOnlineCount(source string, adapter IFetchOnlineCount) error {
	logger.ZInfo("SyncOnlineCount", zap.String("source", source))

	onlineCountList, err := adapter.FetchOnlineCount()
	if err != nil {
		return err
	}

	// 构造批量更新数据
	var updateList []*entities.GameRefresh

	for _, onlineCount := range onlineCountList {
		gameName := onlineCount["gameName"].(string)
		count := int(onlineCount["count"].(float64))

		updateList = append(updateList, &entities.GameRefresh{
			GameCode:    gameName,
			OnlineCount: count,
		})

		logger.ZInfo("SyncOnlineCount",
			zap.String("GameCode", gameName),
			zap.Int("onlineCount", count))
	}

	// 异步持久化到 Redis 和数据库
	go func(list []*entities.GameRefresh) {
		defer utils.PrintPanicStack()
		if err := s.Repo.CacheGameRefreshList(list, s.cacheTTL); err != nil {
			logger.ZError("SyncOnlineCount redis更新失败",
				zap.Error(err),
				zap.String("source", source))
		}

		// if err := s.Repo.SyncOnlineCount2DB(list); err != nil {
		// 	logger.ZError("SyncOnlineCount 数据库更新失败",
		// 		zap.Error(err),
		// 		zap.String("source", source))
		// }
	}(updateList)

	return nil
}
