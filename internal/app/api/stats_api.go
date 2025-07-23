package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var StatsAPISet = wire.NewSet(wire.Struct(new(StatsAPI), "*"))

type StatsAPI struct {
	Srv *service.StatsService
}

func (c *StatsAPI) GetGameStats(ctx *gin.Context) {

	list, err := c.Srv.GetGameStats(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, list)
}

func (c *StatsAPI) GetCategoryStatsList(ctx *gin.Context) {
	var req entities.GetCategoryStatsListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.GetCategoryStatsList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *StatsAPI) GetGameRecordList(ctx *gin.Context) {
	var req entities.GetGameRecordListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.GetGameRecordList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *StatsAPI) GetGamerDailyStatsList(ctx *gin.Context) {
	var req entities.GetGamerDailyStatsListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.GetGamerDailyStatsList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	rsp := buildGamerDailyStatsListRsp(&req)
	ginx.RespSucc(ctx, rsp)
}

func buildGamerDailyStatsListRsp(req *entities.GetGamerDailyStatsListReq) *entities.GetGamerDailyStatsListRsp {
	stat := &entities.GameStats{}
	if stats, ok := req.List.([]*entities.GamerDailyStats); ok {
		for _, s := range stats {
			stat.TotalBetCount += 1
			stat.TotalBetAmount += s.BetAmount
			stat.TotalProfit += s.Profit
		}
	}
	return &entities.GetGamerDailyStatsListRsp{
		Page:     req.Page,
		PageSize: req.PageSize,
		Count:    req.Count,
		List:     req.List,
		Stat:     stat,
	}
}

func (c *StatsAPI) GetUserProfitLeaderboard(ctx *gin.Context) {

	list, err := c.Srv.GetUserProfitLeaderboard(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, list)
}

func (c *StatsAPI) GetProfitLeaderboard(ctx *gin.Context) {

	list, err := c.Srv.GetProfitLeaderboard()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, list)
}
