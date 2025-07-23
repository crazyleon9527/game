package service

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/service/repository"

	"github.com/google/wire"
)

var TransactionServiceSet = wire.NewSet(
	ProvideTransactionService,
)

type TransactionService struct {
	Repo       *repository.TransactionRepository
	UserSrv    *UserService
	apiUrl     string
	appID      string
	appSecret  string
	signSecret string

	walletSrv *WalletService
}

func ProvideTransactionService(repo *repository.TransactionRepository,
	userSrv *UserService,
	walletSrv *WalletService,
) *TransactionService {

	return &TransactionService{
		Repo:      repo,
		UserSrv:   userSrv,
		walletSrv: walletSrv,
	}
}

func (s *TransactionService) GetTransactionList(req *entities.GetTransactionListReq) error {
	// return s.Repo.SearchGameList(req)
	return errors.With("not implement")
}
