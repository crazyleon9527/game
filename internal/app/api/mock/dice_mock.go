package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// DiceGameSet 注入DiceGame
var DiceGameSet = wire.NewSet(wire.Struct(new(DiceGame), "*"))

type DiceGame struct {
}

// DiceGameGetOrderList 获取dice游戏下注列表
// @Summary 获取dice游戏下注列表
// @Description 获取dice游戏下注列表
// @Tags dice游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.DiceGameGetOrderListReq true "params"
// @Success 200 {array} []entities.DiceGameState "成功返回游戏回合"
// @Router /api/dicegame/dice-game-get-order-list [post]
func (c *DiceGame) DiceGameGetOrderList(ctx *gin.Context) {
}

// DiceGamePlaceBet 下注
// @Summary 下注
// @Description 下注
// @Tags dice游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.DiceGamePlaceBetReq true "params"
// @Success 200 {object} entities.DiceGameState "成功返回游戏回合"
// @Router /api/dicegame/dice-game-place-bet [post]
func (c *DiceGame) DiceGamePlaceBet(ctx *gin.Context) {
}

// DiceGameChangeSeed 切换种子
// @Summary 切换种子
// @Description 切换种子
// @Tags dice游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.DiceGameChangeSeedReq true "params"
// @Success 200 {object} entities.DiceGameChangeSeedRsp "成功返回游戏回合"
// @Router /api/dicegame/dice-game-change-seed [post]
func (c *DiceGame) DiceGameChangeSeed(ctx *gin.Context) {
}
