package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// GameSet 注入Game
var GameSet = wire.NewSet(wire.Struct(new(Game), "*"))

type Game struct {
}

// @Tags Game
// @Summary 获取游戏列表
// @Accept  json
// @Produce  json
// @Param req body entities.GetGameListReq true "params"
// @Success 200 {object} entities.Paginator{List=[]entities.Game}
// @Router /api/game/get-game-list [post]
func (a *Game) GetGameList(c *gin.Context) {
}

// @Tags Game
// @Summary 搜索游戏
// @Accept  json
// @Produce  json
// @Param req body entities.SearchGameReq true "params"
// @Success 200 {object} entities.Paginator{List=[]entities.Game}
// @Router /api/game/search [post]
func (a *Game) SearchGame(c *gin.Context) {
}

// @Tags Game
// @Summary 刷新游戏动态数据 ,比如在线人数等等
// @Accept  json
// @Produce  json
// @Success 200 {object} entities.Paginator{List=[]entities.GameRefresh}
// @Router /api/game/refresh-game-list [post]
func (a *Game) RefreshGameList(c *gin.Context) {
}
