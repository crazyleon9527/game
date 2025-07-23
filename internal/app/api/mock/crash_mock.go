package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// CrashGameSet 注入CrashGame
var CrashGameSet = wire.NewSet(wire.Struct(new(CrashGame), "*"))

type CrashGame struct {
}

// WsHandler godoc
// @Summary WebSocket 连接建立
// @Description 与crash服务器建立 WebSocket new WebSocket("ws://url:port/api/crashgame/ws?token=xxx");
// @Tags crash游戏
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {string} string "WebSocket connection established"
// @Success 200 {object} entities.GetCrashGameRoundRsp "成功返回游戏回合"
// @Success 200 {object} entities.CrashGameOrderNotify "成功返回游戏回合下注"
// @Router /api/crashgame/ws [get]
func (c *CrashGame) WsHandler(ctx *gin.Context) {
}

// GetCrashGameRound 获取crash游戏状态
// @Summary 获取crash游戏状态
// @Description 获取crash游戏状态
// @Tags crash游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.GetCrashGameRoundReq true "params"
// @Success 200 {object} entities.GetCrashGameRoundRsp "成功返回游戏回合"
// @Router /api/crashgame/get-crash-game-round [post]
func (c *CrashGame) GetCrashGameRound(ctx *gin.Context) {
}

// GetCrashGameRoundList 获取crash游戏历史记录
// @Summary 获取crash游戏历史记录
// @Description 获取crash游戏历史记录
// @Tags crash游戏
// @Produce json
// @Param req body entities.GetCrashGameRoundListReq true "params"
// @Success 200 {object} entities.GetCrashGameRoundListRsp "成功返回游戏回合列表"
// @Router /api/crashgame/get-crash-game-round-list [post]
func (c *CrashGame) GetCrashGameRoundList(ctx *gin.Context) {
}

// GetCrashGameRoundOrderList 获取crash游戏下注列表
// @Summary 获取crash游戏下注列表
// @Description 获取crash游戏下注列表
// @Tags crash游戏
// @Produce json
// @Param req body entities.GetCrashGameRoundOrderListReq true "params"
// @Success 200 {array} []entities.CrashGameOrder "成功返回游戏回合下注列表"
// @Router /api/crashgame/get-crash-game-round-order-list [post]
func (c *CrashGame) GetCrashGameRoundOrderList(ctx *gin.Context) {
}

// PlaceCrashGameBet 下注crash游戏
// @Summary 下注crash游戏
// @Description 下注crash游戏
// @Tags crash游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.PlaceCrashGameBetReq true "params"
// @Success 200 {object} entities.CrashGameOrder "成功返回游戏订单"
// @Router /api/crashgame/place-crash-game-bet [post]
func (c *CrashGame) PlaceCrashGameBet(ctx *gin.Context) {
}

// CancelCrashGameBet 取消crash游戏下注
// @Summary 取消crash游戏下注
// @Description 取消crash游戏下注
// @Tags crash游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.CancelCrashGameBetReq true "params"
// @Success 200
// @Router /api/crashgame/cancel-crash-game-bet [post]
func (c *CrashGame) CancelCrashGameBet(ctx *gin.Context) {
}

// EscapeCrashGameBet 逃跑crash游戏
// @Summary 逃跑crash游戏
// @Description 逃跑crash游戏
// @Tags crash游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.EscapeCrashGameBetReq true "params"
// @Success 200 {object} entities.CrashGameOrder "成功返回游戏订单"
// @Router /api/crashgame/escape-crash-game-bet [post]
func (c *CrashGame) EscapeCrashGameBet(ctx *gin.Context) {
}

// // GetCrashAutoBet 获取自动下注crash游戏
// // @Summary 获取自动下注crash游戏
// // @Description 获取自动下注crash游戏
// // @Tags crash游戏
// // @Produce json
// // @Security ApiKeyAuth
// // @Param req body entities.GetCrashAutoBetReq true "params"
// // @Success 200 {object} entities.CrashAutoBet "成功返回游戏订单"
// // @Router /api/crashgame/get-auto-crash-game-bet [post]
// func (c *CrashGame) GetCrashAutoBet(ctx *gin.Context) {
// }

// // PlaceCrashAutoBet 自动下注crash游戏
// // @Summary 自动下注crash游戏
// // @Description 自动下注crash游戏
// // @Tags crash游戏
// // @Produce json
// // @Security ApiKeyAuth
// // @Param req body entities.PlaceCrashAutoBetReq true "params"
// // @Success 200 {object} entities.CrashAutoBet "成功返回游戏订单"
// // @Router /api/crashgame/place-auto-crash-game-bet [post]
// func (c *CrashGame) PlaceCrashAutoBet(ctx *gin.Context) {
// }

// // CancelCrashAutoBet 取消自动下注crash游戏
// // @Summary 取消自动下注crash游戏
// // @Description 取消自动下注crash游戏
// // @Tags crash游戏
// // @Produce json
// // @Security ApiKeyAuth
// // @Param req body entities.CancelCrashAutoBetReq true "params"
// // @Success 200
// // @Router /api/crashgame/cancel-auto-crash-game-bet [post]
// func (c *CrashGame) CancelCrashAutoBet(ctx *gin.Context) {
// }

// GetUserCrashGameOrder 获取用户crash游戏下注
// @Summary 获取用户crash游戏下注
// @Description 获取用户crash游戏下注
// @Tags crash游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.GetUserCrashGameOrderReq true "params"
// @Success 200 {array} []entities.CrashGameOrder "成功返回用户游戏回合下注"
// @Router /api/crashgame/get-user-crash-game-order [post]
func (c *CrashGame) GetUserCrashGameOrder(ctx *gin.Context) {
}

// GetUserCrashGameOrderList 获取用户crash游戏下注列表
// @Summary 获取用户crash游戏下注列表
// @Description 获取用户crash游戏下注列表
// @Tags crash游戏
// @Produce json
// @Security ApiKeyAuth
// @Param req body entities.GetUserCrashGameOrderListReq true "params"
// @Success 200 {object} entities.GetUserCrashGameOrderListRsp "成功返回用户游戏回合下注列表"
// @Router /api/crashgame/get-user-crash-game-order-list [post]
func (c *CrashGame) GetUserCrashGameOrderList(ctx *gin.Context) {
}
