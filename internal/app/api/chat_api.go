package api

import (
	"net/http"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"
	"rk-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var ChatAPISet = wire.NewSet(wire.Struct(new(ChatAPI), "*"))

type ChatAPI struct {
	Srv *service.ChatService
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (c *ChatAPI) WsHandler(ctx *gin.Context) {
	uid := ginx.Mine(ctx)
	logger.ZInfo("WsHandler", zap.Uint("uid", uid))
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	client := c.Srv.Connect(uid, conn)
	logger.ZInfo("client connected")
	// 获取离线消息
	messages := c.Srv.GetOfflineMessages(uid)
	for _, msg := range messages {
		client.Send <- []byte(msg)
	}

	// 启动客户端
	client.Start()
}

func (c *ChatAPI) GetChannelList(ctx *gin.Context) {

	channels, err := c.Srv.GetChannelList()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, channels)
}

func (c *ChatAPI) JoinChannel(ctx *gin.Context) {
	var req entities.JoinChannelReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Srv.JoinChannel(ginx.Mine(ctx), req.Channel)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, nil)
}
func (c *ChatAPI) SendMessage(ctx *gin.Context) {
	var msg entities.ChatMessageReq
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	chatMessage := entities.ChatMessage{
		SenderID:   ginx.Mine(ctx),
		ReceiverID: msg.ReceiverID,
		Channel:    msg.Channel,
		Content:    msg.Content,
		Type:       msg.Type,
	}

	err := c.Srv.SendMessage(&chatMessage)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *ChatAPI) GetHistory(ctx *gin.Context) {
	var req entities.GetMessageHistoryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.GetMessageHistory(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *ChatAPI) UploadImage(ctx *gin.Context) {

}
