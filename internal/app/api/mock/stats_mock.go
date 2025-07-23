package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// GameSet 注入Game
var StatsSet = wire.NewSet(wire.Struct(new(Stats), "*"))

type Stats struct {
}

// @Summary 获取游戏统计数据
// @Description 获取游戏统计数据
// @Tags Stats[统计]
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} entities.GameStats
// @Router /api/stats/get-game-stats [post]
func (c *Stats) GetGameStats(ctx *gin.Context) {

}

// @Summary 获取游戏类比统计列表数据
// @Tags Stats[统计]
// @Description 获取游戏类比统计列表数据
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.GetCategoryStatsListReq true "查询条件"
// @Success 200 {array} entities.GameCategoryStats
// @Router /api/stats/get-category-stats-list [post]
func (c *Stats) GetCategoryStatsList(ctx *gin.Context) {

}

// @Summary 获取游戏记录详情列表数据
// @Tags Stats[统计]
// @Description 获取游戏记录详情列表数据
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.GetGameRecordListReq true "查询条件"
// @Success 200 {array} entities.GameRecord
// @Router /api/stats/get-game-record-list [post]
func (c *Stats) GetGameRecordList(ctx *gin.Context) {

}

// @Summary 玩家每日统计数据列表
// @Tags Stats[统计]
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.GetGamerDailyStatsListReq true "查询条件"
// @Success 200 {object} entities.GetGamerDailyStatsListRsp
// @Router /api/stats/get-gamer-daily-stats-list [post]
func (c *Stats) GetGamerDailyStatsList(ctx *gin.Context) {

}

// @Summary 获取某玩家游戏盈利排行
// @Tags Stats[统计]
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} entities.GameProfitStats
// @Router /api/stats/get-gamer-profit-leaderboard [post]
func (c *Stats) GetUserProfitLeaderboard(ctx *gin.Context) {

}

// @Summary 获取所有玩家游戏盈利排行
// @Tags Stats[统计]
// @Accept json
// @Produce json
// @Success 200 {array} entities.GameProfitStats
// @Router /api/stats/get-profit-leaderboard [post]
func (c *Stats) GetProfitLeaderboard(ctx *gin.Context) {

}
