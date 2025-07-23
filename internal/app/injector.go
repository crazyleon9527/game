package internal

import (
	game "rk-api/internal/app/game/rg"
	"rk-api/internal/app/service/async"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var InjectorSet = wire.NewSet(wire.Struct(new(Injector), "*"))

type Injector struct {
	Engine  *gin.Engine
	Service async.IAsyncService
	Nine    game.INine
	Wingo   game.IWingo
}
