package repository

import (
	"context"
	"fmt"
	"rk-api/internal/app/entities"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ChatRepositorySet = wire.NewSet(wire.Struct(new(ChatRepository), "*"))

type ChatRepository struct {
	DB  *gorm.DB
	RDS redis.UniversalClient
}

func (r *ChatRepository) CreateChatMessage(entity *entities.ChatMessage) error {
	return r.DB.Create(entity).Error
}

// 发布频道消息
func (r *ChatRepository) PublishChannelMessage(channel string, msg []byte) {
	r.RDS.Publish(context.Background(), fmt.Sprintf("channel:%s", channel), msg)
}

func (c *ChatRepository) GetOfflineMessages(uid uint) []string {
	// 获取并发送离线消息
	key := fmt.Sprintf("private:%d", uid)
	messages, _ := c.RDS.LRange(context.Background(), key, 0, -1).Result()
	c.RDS.Del(context.Background(), key)
	return messages
}

// 存储离线消息
func (r *ChatRepository) PushPrivateMessage(receiverID uint, msg []byte) {
	key := fmt.Sprintf("private:%d", receiverID)
	r.RDS.RPush(context.Background(), key, msg)
}

func (r *ChatRepository) GetChannelList() ([]*entities.ChatChannel, error) {
	var channels []*entities.ChatChannel
	err := r.DB.Where("active = ? ", 1).Find(&channels).Error
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// 根据channel获取聊天频道
func (r *ChatRepository) GetChannel(channel string) (*entities.ChatChannel, error) {
	var channelEntity entities.ChatChannel
	err := r.DB.Where("name = ? ", channel).First(&channelEntity).Error
	if err != nil {
		return nil, err
	}
	return &channelEntity, nil
}

func (r *ChatRepository) GetMessageHistory(param *entities.GetMessageHistoryReq) error {
	var tx *gorm.DB = r.DB
	if param.Channel != "" {
		tx = tx.Where("channel = ?", param.Channel).Order("created_at desc")
	}
	param.List = make([]*entities.ChatMessage, 0)
	return param.Paginate(tx)
}
