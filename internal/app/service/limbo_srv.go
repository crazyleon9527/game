package service

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/repository"

	"github.com/google/wire"
)

var LimboGameServiceSet = wire.NewSet(
	ProvideLimboGameService,
)

type LimboGameService struct {
	Repo      *repository.LimboGameRepository
	UserSrv   *UserService
	WalletSrv *WalletService
}

func ProvideLimboGameService(
	repo *repository.LimboGameRepository,
	userSrv *UserService,
	walletSrv *WalletService,
) *LimboGameService {
	service := &LimboGameService{
		Repo:      repo,
		UserSrv:   userSrv,
		WalletSrv: walletSrv,
	}
	return service
}

// GetUserLimboGameOrder retrieves the user's Limbo game order.
func (s *LimboGameService) GetUserLimboGameOrder(uid uint) (*entities.LimboGameOrder, error) {
	return s.Repo.GetUserLimboGameOrder(uid)
}

// GetUserLimboGameOrderList retrieves the user's Limbo game order list.
func (s *LimboGameService) GetUserLimboGameOrderList(uid uint) ([]*entities.LimboGameOrder, error) {
	return s.Repo.GetUserLimboGameOrderList(uid)
}

// UpdateLimboGameOrder updates a Limbo game order.
func (s *LimboGameService) UpdateLimboGameOrder(order *entities.LimboGameOrder) error {
	return s.Repo.UpdateLimboGameOrder(order)
}

// CreateLimboGameOrder creates a new Limbo game order.
func (s *LimboGameService) CreateLimboGameOrder(order *entities.LimboGameOrder) error {
	return s.Repo.CreateLimboGameOrder(order)
}

// PlaceOrder processes a Limbo game order.
func (s *LimboGameService) PlaceOrder(order *entities.LimboGameOrder) error {
	// Implement business logic for placing a Limbo game order
	return nil
}

// SettleOrder settles a Limbo game order.
func (s *LimboGameService) SettleOrder(order *entities.LimboGameOrder) error {
	// Implement business logic for settling a Limbo game order
	return nil
}
