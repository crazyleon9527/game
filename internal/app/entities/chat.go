package entities

// 消息模型
type ChatMessage struct {
	BaseModel
	SenderID   uint   `json:"senderId"`
	ReceiverID uint   `json:"receiverId"` // 单聊时使用
	Channel    string `json:"channel"`    // 群聊时使用
	Content    string `json:"content"`    // 消息内容
	Type       string // "text", "image"
}

type ChatChannel struct {
	BaseModel
	Name     string `gorm:"uniqueIndex" redis:"name"  json:"name"`
	Language string `json:"language"` // "zh-CN", "en-US"
	Active   uint8  `json:"active"`   // 是否活跃
}

type GetMessageHistoryReq struct {
	Paginator
	Channel string `json:"channel"`
}

type ChatMessageReq struct {
	ReceiverID uint   // 单聊时使用
	Channel    string `json:"channel"` // 群聊时使用
	Content    string `json:"content"` // 消息内容
	Type       string `json:"type"`    // "text", "image"
}

type JoinChannelReq struct {
	Channel string `json:"channel"`
}
