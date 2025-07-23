package service

import (
	"errors"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/repository"
	"time"

	"github.com/google/wire"
	"github.com/orca-zhang/ecache"
)

var RealServiceSet = wire.NewSet(
	ProvideRealService,
)

type RealService struct {
	Repo      *repository.RealRepository
	UserSrv   *UserService
	platCache *ecache.Cache
}

func ProvideRealService(
	repo *repository.RealRepository,
	userSrv *UserService,

) *RealService {
	service := &RealService{
		Repo:      repo,
		UserSrv:   userSrv,
		platCache: ecache.NewLRUCache(1, 1, 10*time.Minute), //初始化缓存
	}
	return service
}

func (s *RealService) CreateRealAuth(auth *entities.RealAuth) error {
	// 这里可以添加一些额外的业务逻辑，例如验证身份证号、手机号等
	if auth.RealName == "" || auth.IDCard == "" {
		return errors.New("real name and ID card cannot be empty")
	}
	return s.Repo.CreateRealAuth(auth)
}

func (s *RealService) GetRealAuthByUID(uid uint) (*entities.RealAuth, error) {
	return s.Repo.GetRealAuthByUID(uid)
}

func (s *RealService) UpdateRealAuth(auth *entities.UpdateRealNameAuthReq) error {
	return s.Repo.UpdateRealAuth(&entities.RealAuth{
		UID:      auth.UID,
		RealName: auth.RealName,
		IDCard:   auth.IDCard,
	})
}
