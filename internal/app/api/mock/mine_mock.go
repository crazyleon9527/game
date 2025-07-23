package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// MineGameSet 注入MineGame
var MineGameSet = wire.NewSet(wire.Struct(new(MineGame), "*"))

type MineGame struct {
}

// MineGameGetState 获取mine游戏状态
// @Summary 获取mine游戏状态
// @Description 获取mine游戏状态
// @Tags mine游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.MineGameGetStateReq true "params"
// @Success 200 {object} entities.MineGameState "成功返回游戏回合"
// @Router /api/minegame/mine-game-get-state [post]
func (c *MineGame) MineGameGetState(ctx *gin.Context) {
}

// MineGameGetOrderList 获取mine游戏下注列表
// @Summary 获取mine游戏下注列表
// @Description 获取mine游戏下注列表
// @Tags mine游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.MineGameGetOrderListReq true "params"
// @Success 200 {array} []entities.MineGameState "成功返回游戏回合"
// @Router /api/minegame/mine-game-get-order-list [post]
func (c *MineGame) MineGameGetOrderList(ctx *gin.Context) {
}

// MineGamePlaceBet 下注
// @Summary 下注
// @Description 下注
// @Tags mine游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.MineGamePlaceBetReq true "params"
// @Success 200 {object} entities.MineGameState "成功返回游戏回合"
// @Router /api/minegame/mine-game-place-bet [post]
func (c *MineGame) MineGamePlaceBet(ctx *gin.Context) {
}

// MineGameOpenPosition 开启位置
// @Summary 开启位置
// @Description 开启位置
// @Tags mine游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.MineGameOpenPositionReq true "params"
// @Success 200 {object} entities.MineGameState "成功返回游戏回合"
// @Router /api/minegame/mine-game-open-position [post]
func (c *MineGame) MineGameOpenPosition(ctx *gin.Context) {
}

// MineGameCashout 兑现
// @Summary 兑现
// @Description 兑现
// @Tags mine游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.MineGameCashoutReq true "params"
// @Success 200 {object} entities.MineGameState "成功返回游戏回合"
// @Router /api/minegame/mine-game-cashout [post]
func (c *MineGame) MineGameCashout(ctx *gin.Context) {
}

// MineGameChangeSeed 切换种子
// @Summary 切换种子
// @Description 切换种子
// @Tags mine游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.MineGameChangeSeedReq true "params"
// @Success 200 {object} entities.MineGameChangeSeedRsp "成功返回游戏回合"
// @Router /api/minegame/mine-game-change-seed [post]
func (c *MineGame) MineGameChangeSeed(ctx *gin.Context) {
}
