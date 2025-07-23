package service

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/repository"

	"github.com/google/wire"
)

var FinancialServiceSet = wire.NewSet(
	ProvideFinancialService,
)

type FinancialService struct {
	Repo              *repository.FinancialRepository
	UserSrv           *UserService
	AgentSrv          *AgentService
	StateSrv          *StateService
	InFinancialReturn bool
}

func ProvideFinancialService(
	repo *repository.FinancialRepository,
) *FinancialService {
	service := &FinancialService{
		Repo: repo,
	}
	return service
}

func (s *FinancialService) GetSummary(uid uint) (*entities.FinancialSummary, error) {
	return s.Repo.GetSummary(uid)
}
func (s *FinancialService) CreateFinancialSummary(summary *entities.FinancialSummary) error {
	return s.Repo.CreateFinancialSummary(summary)
}
