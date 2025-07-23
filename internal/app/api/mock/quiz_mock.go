package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// QuizSet 注入Quiz
var QuizSet = wire.NewSet(wire.Struct(new(Quiz), "*"))

type Quiz struct {
}

// @Tags Quiz
// @Summary 获取竞猜信息
// @Description 获取竞猜信息
// @Accept  json
// @Produce  json
// @Success 200 {object} entities.QuizInfoRsp
// @Router /api/quiz/get-quiz-info [post]
func (a *Quiz) GetQuizInfo(c *gin.Context) {
}

// @Tags Quiz
// @Summary 获取竞猜信息
// @Description 获取竞猜信息
// @Accept  json
// @Produce  json
// @Param request body entities.QuizListReq true "查询条件"
// @Success 200 {object} []entities.QuizInfoRsp
// @Router /api/quiz/get-quiz-list [post]
func (a *Quiz) GetQuizList(c *gin.Context) {
}

// @Tags Quiz
// @Summary 竞猜购买
// @Description 竞猜购买
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param request body entities.QuizBuyReq true "params"
// @Success 200 {object} entities.QuizBuyRecord
// @Router /api/quiz/quiz-buy [post]
func (a *Quiz) QuizBuy(c *gin.Context) {
}

// @Tags Quiz
// @Summary 获取竞猜购买记录
// @Description 获取竞猜购买记录
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param request body entities.QuizBuyRecordReq true "查询条件"
// @Success 200 {object} []entities.QuizBuyRecord
// @Router /api/quiz/get-quiz-buy-record [post]
func (a *Quiz) GetQuizBuyRecord(c *gin.Context) {
}

// @Tags Quiz
// @Summary 竞猜价格历史
// @Description 竞猜价格历史
// @Accept  json
// @Produce  json
// @Param request body entities.QuizPricesHistoryReq true "查询条件"
// @Success 200 {object} entities.QuizPricesHistoryRsp
// @Router /api/quiz/get-quiz-prices-history [post]
func (a *Quiz) GetQuizPricesHistory(c *gin.Context) {
}

// @Tags Quiz
// @Summary 竞猜市场价格历史
// @Description 竞猜市场价格历史
// @Accept  json
// @Produce  json
// @Param request body entities.QuizMarketPricesHistoryReq true "查询条件"
// @Success 200 {object} entities.QuizPricesHistoryRspItem
// @Router /api/quiz/get-quiz-market-prices-history [post]
func (a *Quiz) GetQuizMarketPricesHistory(c *gin.Context) {
}
