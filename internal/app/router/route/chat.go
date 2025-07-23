package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterChatRoutes(r *gin.RouterGroup, chatAPI *api.ChatAPI) {
	chat := r.Group("/chat")
	{
		chat.POST("/get-channel-list", middleware.JWTMiddleware(), chatAPI.GetChannelList)
		chat.POST("/join-channel", middleware.JWTMiddleware(), chatAPI.JoinChannel)
		chat.POST("/send-message", middleware.JWTMiddleware(), chatAPI.SendMessage)
		chat.POST("/get-history", middleware.JWTMiddleware(), chatAPI.GetHistory)
		chat.POST("/upload", middleware.JWTMiddleware(), chatAPI.UploadImage)
		chat.GET("/ws", middleware.JWTMiddleware(), chatAPI.WsHandler)
	}
}
