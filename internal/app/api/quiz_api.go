package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"
	"rk-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var QuizAPISet = wire.NewSet(wire.Struct(new(QuizAPI), "*"))

type QuizAPI struct {
	Srv *service.QuizService
}

func (c *QuizAPI) GetQuizInfo(ctx *gin.Context) {
	info, err := c.Srv.GetQuizInfo()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, info)
}

func (c *QuizAPI) GetQuizList(ctx *gin.Context) {
	var req entities.QuizListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.GetQuizList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *QuizAPI) QuizBuy(ctx *gin.Context) {
	var req entities.QuizBuyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		logger.Errorf("QuizBuy ShouldBindJSON err, err - %s", err)
		return
	}
	req.UID = ginx.Mine(ctx)
	record, err := c.Srv.QuizBuy(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		logger.Errorf("QuizBuy QuizBuy err, err - %s", err)
		return
	}
	ginx.RespSucc(ctx, record)
}

func (c *QuizAPI) GetQuizBuyRecord(ctx *gin.Context) {
	var req entities.QuizBuyRecordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.GetQuizBuyRecord(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

// GetQuizPricesHistory
func (c *QuizAPI) GetQuizPricesHistory(ctx *gin.Context) {
	var req entities.QuizPricesHistoryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	rsp, err := c.Srv.GetQuizPricesHistory(req.EventID)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// GetQuizMarketPricesHistory
func (c *QuizAPI) GetQuizMarketPricesHistory(ctx *gin.Context) {
	var req entities.QuizMarketPricesHistoryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	rsp, err := c.Srv.GetQuizMarketPricesHistory(req.EventID, req.MarketID)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}
