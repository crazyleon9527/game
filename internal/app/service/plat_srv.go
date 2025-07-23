package service

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/repository"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
)

var PlatServiceSet = wire.NewSet(
	ProvidePlatService,
)

type PlatService struct {
	Repo      *repository.PlatRepository
	UserSrv   *UserService
	platCache *ecache.Cache
}

func ProvidePlatService(
	repo *repository.PlatRepository,
	userSrv *UserService,

) *PlatService {
	service := &PlatService{
		Repo:      repo,
		UserSrv:   userSrv,
		platCache: ecache.NewLRUCache(1, 1, 10*time.Minute), //初始化缓存
	}
	return service
}

// 获取返利列表
func (s *PlatService) GetPlatSetting() (*entities.PlatSetting, error) {
	// 从缓存中获取数据
	if val, ok := s.platCache.Get("plat_setting"); ok {
		return val.(*entities.PlatSetting), nil
	}
	// 从数据库中获取数据
	setting, err := s.Repo.GetPlatSetting()
	if err != nil {
		return nil, err
	}
	// 缓存数据
	s.platCache.Put("plat_setting", setting)
	return setting, nil
}
