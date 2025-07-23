package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// ChatSet 注入Chat
var ChatSet = wire.NewSet(wire.Struct(new(Chat), "*"))

type Chat struct {
}

// WsHandler godoc
// @Summary WebSocket 连接建立
// @Description 与聊天服务器建立 WebSocket new WebSocket("ws://url:port/api/chat/ws?token=xxx");
// @Tags Chat
// @Accept json
// @Produce json
// @Success 200 {string} string "WebSocket connection established"
// @Router /api/chat/ws [get]
func (c *Chat) WsHandler(ctx *gin.Context) {

}

// GetChannelList godoc
// @Summary 获取所有可用聊天频道列表
// @Description 获取所有可用的聊天频道列表
// @Tags Chat
// @Accept json
// @Produce json
// @Success 200 {array} entities.ChatChannel
// @Router  /api/chat/get-channel-list [post]
func (c *Chat) GetChannelList(ctx *gin.Context) {
	// 实现代码
}

// @Summary 发送聊天信息
// @Description 发送聊天信息
// @Tags Chat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param message body entities.ChatMessage true "Message"
// @Success 200
// @Router /api/chat/send-message [post]
func (c *Chat) SendMessage(ctx *gin.Context) {
	// 实现代码
}

// UploadImage godoc
// @Summary 上传聊天图片
// @Description 上传并保存聊天图片
// @Tags Chat
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param image formData file true "Image File"
// @Success 200
// @Router /api/chat/upload-image [post]
func (c *Chat) UploadImage(ctx *gin.Context) {

}

// GetHistory godoc
// @Summary 获取聊天频道历史记录
// @Description 获取指定频道的历史聊天记录
// @Tags Chat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.GetMessageHistoryReq true "查询条件"
// @Success 200 {array} entities.ChatMessage
// @Router /api/chat/get-history [post]
func (c *Chat) GetHistory(ctx *gin.Context) {

}

// JoinChannel godoc
// @Summary 加入一个聊天频道
// @Description 用户加入指定的聊天频道
// @Tags Chat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.JoinChannelReq true "查询条件"
// @Success 200
// @Router /api/chat/join-channel [post]
func (c *Chat) JoinChannel(ctx *gin.Context) {

}
