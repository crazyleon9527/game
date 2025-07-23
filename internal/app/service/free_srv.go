package service

import (
	"fmt"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/service/repository"

	"github.com/google/wire"
)

var FreeServiceSet = wire.NewSet(
	ProvideFreeService,
)

type FreeService struct {
	Repo    *repository.FreeRepository
	UserSrv *UserService
}

func ProvideFreeService(repo *repository.FreeRepository,
	userSrv *UserService,
) *FreeService {
	return &FreeService{
		Repo:    repo,
		UserSrv: userSrv,
	}
}

func (s *FreeService) GenerateFreeCard(uid uint, amount int64) (*entities.FreeCard, error) {
	if amount <= 400 {
		return nil, fmt.Errorf("充值金额必须大于400才能获得免单卡")
	}
	var freeCard entities.FreeCard
	freeCard.UID = uid
	freeCard.Amount = float64(amount)
	freeCard.Used = false
	err := s.Repo.CreateFreeCard(&freeCard)
	if err != nil {
		return nil, err
	}
	return &freeCard, nil
}

func (s *FreeService) GetAvailableFreeCard(uid uint, amount int64) (*entities.FreeCard, error) {
	return s.Repo.GetAvailableFreeCard(uid, amount)
}

// UseFreeCard 使用免单卡支付
func (s *FreeService) UseFreeCard(uid uint, id uint) (*entities.FreeCard, error) {
	// 获取金额大于等于消费金额的最小免单卡
	card, err := s.Repo.GetFreeCard(id)
	if err != nil {
		return nil, err
	}
	if card.UID != uid {
		return nil, errors.With("free card not belong to this user")
	}

	if card.Used {
		return nil, errors.With("free card has been used")
	}
	// 标记免单卡为已使用
	card.Used = true
	if err := s.Repo.Update(card); err != nil {
		return nil, err
	}

	return card, nil
}
