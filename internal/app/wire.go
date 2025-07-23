//go:build wireinject
// +build wireinject

package internal

// The build tag makes sure the stub is not built in the final build.
import (
	"rk-api/internal/app/api"
	"rk-api/internal/app/router"
	"rk-api/internal/app/service"
	"rk-api/internal/app/service/repository"

	"github.com/google/wire"
)

func BuildInjector() (*Injector, func(), error) {
	wire.Build(
		ProviderSet,
		repository.RepoSet,
		service.SrvSet,
		api.APISet,
		router.RouterSet,
		InitGinEngine,
		InjectorSet,
	)
	return new(Injector), nil, nil
}
