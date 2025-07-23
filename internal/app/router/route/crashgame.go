package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterCrashGameRoutes(r *gin.RouterGroup, crashAPI *api.CrashGameAPI) {
	crash := r.Group("/crashgame")
	{
		crash.GET("/ws", middleware.JWTMiddleware(), crashAPI.WsHandler)
		crash.POST("/get-crash-game-round", middleware.JWTMiddleware(), crashAPI.GetCrashGameRound)
		crash.POST("/get-crash-game-round-list", crashAPI.GetCrashGameRoundList)
		crash.POST("/get-crash-game-round-order-list", crashAPI.GetCrashGameRoundOrderList)
		crash.POST("/place-crash-game-bet", middleware.JWTMiddleware(), crashAPI.PlaceCrashGameBet)
		crash.POST("/cancel-crash-game-bet", middleware.JWTMiddleware(), crashAPI.CancelCrashGameBet)
		crash.POST("/escape-crash-game-bet", middleware.JWTMiddleware(), crashAPI.EscapeCrashGameBet)
		crash.POST("/get-user-crash-game-order", middleware.JWTMiddleware(), crashAPI.GetUserCrashGameOrder)
		crash.POST("/get-user-crash-game-order-list", middleware.JWTMiddleware(), crashAPI.GetUserCrashGameOrderList)
	}
}
