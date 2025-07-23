package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// SDGameSet 注入SDGame
var SDGameSet = wire.NewSet(wire.Struct(new(SDGame), "*"))

type SDGame struct {
}

// PlaceSDBet 单双投注
// @Summary 单双投注
// @Tags SDGame
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.HashSDBetRequest true "params"
// @Success 200
// @Router /api/sdgame/place-sd-bet [post]
func (a *SDGame) PlaceSDBet(c *gin.Context) {
}

// GetSDGameState 获取单双游戏状态
// @Summary 获取单双游戏房间状态
// @Description 通过 betType 获取指定房间的游戏状态信息
// @Tags SDGame
// @Produce json
// @Param betType query int false "房间类型 (1 初级 2 中级 3 高级)" default(1)
// @Success 200 {object} hash.GameState "成功返回游戏状态"
// @Router /api/sdgame/get-sd-game-state [post]
func (c *SDGame) GetSDGameState(ctx *gin.Context) {
}
