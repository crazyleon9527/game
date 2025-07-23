package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterQuizRoutes(r *gin.RouterGroup, quizAPI *api.QuizAPI) {
	quiz := r.Group("/quiz")
	{
		quiz.POST("/get-quiz-info", quizAPI.GetQuizInfo)
		quiz.POST("/get-quiz-list", quizAPI.GetQuizList)
		quiz.POST("/quiz-buy", middleware.JWTMiddleware(), quizAPI.QuizBuy)
		quiz.POST("/get-quiz-buy-record", middleware.JWTMiddleware(), quizAPI.GetQuizBuyRecord)
		quiz.POST("/get-quiz-prices-history", quizAPI.GetQuizPricesHistory)
		quiz.POST("/get-quiz-market-prices-history", quizAPI.GetQuizMarketPricesHistory)
	}
}
