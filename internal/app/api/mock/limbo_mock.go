package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// LimboGameSet 注入LimboGame
var LimboGameSet = wire.NewSet(wire.Struct(new(LimboGame), "*"))

type LimboGame struct {
}

// LimboGameGetOrderList 获取Limbo游戏下注列表
// @Summary 获取Limbo游戏下注列表
// @Description 获取Limbo游戏下注列表
// @Tags limbo游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.LimboGameGetOrderListReq true "params"
// @Success 200 {array} []entities.LimboGameState "成功返回游戏回合"
// @Router /api/limbogame/limbo-game-get-order-list [post]
func (c *LimboGame) LimboGameGetOrderList(ctx *gin.Context) {
}

// LimboGamePlaceBet 下注
// @Summary 下注
// @Description 下注
// @Tags limbo游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.LimboGamePlaceBetReq true "params"
// @Success 200 {object} entities.LimboGameState "成功返回游戏回合"
// @Router /api/limbogame/limbo-game-place-bet [post]
func (c *LimboGame) LimboGamePlaceBet(ctx *gin.Context) {
}

// LimboGameChangeSeed 切换种子
// @Summary 切换种子
// @Description 切换种子
// @Tags limbo游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.LimboGameChangeSeedReq true "params"
// @Success 200 {object} entities.LimboGameChangeSeedRsp "成功返回游戏回合"
// @Router /api/limbogame/limbo-game-change-seed [post]
func (c *LimboGame) LimboGameChangeSeed(ctx *gin.Context) {
}