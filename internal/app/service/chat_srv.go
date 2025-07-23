package service

import (
	"context"
	"encoding/json"
	"rk-api/internal/app/chat"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/service/repository"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"strings"

	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var ChatServiceSet = wire.NewSet(
	ProvideChatService,
)

type ChatService struct {
	Repo    *repository.ChatRepository
	UserSrv *UserService
	hub     *chat.Hub
}

func ProvideChatService(
	repo *repository.ChatRepository,
	userSrv *UserService,
) *ChatService {
	hub := chat.NewHub()
	service := &ChatService{
		Repo:    repo,
		UserSrv: userSrv,
		hub:     hub,
	}
	go hub.Run()
	service.StartMessageDispatcher()
	return service
}

func (s *ChatService) Connect(uid uint, conn *websocket.Conn) *chat.Client {
	client := &chat.Client{
		UID:       uid,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Channels:  make(map[string]struct{}),
		Hub:       s.hub,
		Processor: s,
	}
	s.hub.Register <- client
	return client
}

/**
 * 从websocket连接中读取消息并处理消息
 * 暂时不需要
 */
func (s *ChatService) ProcessMessage(rawMsg []byte) {
	var msg entities.ChatMessage
	json.Unmarshal(rawMsg, &msg)

	s.SendMessage(&msg)
}

/**
 * 直接发送消息
 */
func (s *ChatService) SendMessage(msg *entities.ChatMessage) error {
	_, err := s.UserSrv.GetUserByUID(msg.SenderID)
	if err != nil {
		return err
	}
	if !s.hub.IsSubscribed(msg.SenderID, msg.Channel) {
		return errors.WithCode(errors.ChatChannelNotSubscribed)
	}
	// logger.ZInfo("send message", zap.String("sender", fmt.Sprintf("%d", msg.SenderID)), zap.String("receiver", fmt.Sprintf("%d", msg.ReceiverID)), zap.String("channel", msg.Channel), zap.String("content", msg.Content))
	//判断条件
	go s.Repo.CreateChatMessage(msg) //存储到数据库
	if msg.Channel != "" {
		// 频道消息
		s.Repo.PublishChannelMessage(msg.Channel, []byte(msg.Content))
	} else {
		s.Repo.PushPrivateMessage(msg.ReceiverID, []byte(msg.Content))
	}
	return nil
}

// 统一的消息分发循环
func (s *ChatService) StartMessageDispatcher() {
	logger.Info("start message dispatcher")
	go func() {
		defer utils.PrintPanicStack()
		ctx := context.Background()
		pubsub := s.Repo.RDS.PSubscribe(ctx, "channel:*")
		defer pubsub.Close()

		ch := pubsub.Channel()
		for msg := range ch {
			// 解析频道ID
			var channel = strings.TrimPrefix(msg.Channel, "channel:")

			logger.ZInfo("receive message from channel", zap.String("channel", channel), zap.String("message", msg.Payload))

			s.hub.Broadcast <- &chat.Broadcast{
				Channel: channel,
				Message: []byte(msg.Payload),
			}
		}
	}()
}

// 加入频道（线程安全）
func (s *ChatService) JoinChannel(uid uint, channel string) error {
	// logger.ZInfo("join channel", zap.String("uid", fmt.Sprintf("%d", uid)), zap.String("channel", channel))
	chatChannel, err := s.Repo.GetChannel(channel)
	if err != nil {
		logger.ZError("get channel error", zap.Error(err))
		return errors.With("channel not found")
	}
	if chatChannel.Active != 1 {
		return errors.With("channel is closed")
	}
	if !s.hub.IsSubscribed(uid, channel) {
		// logger.ZInfo("subscribe channel", zap.String("uid", fmt.Sprintf("%d", uid)), zap.String("channel", channel))
		s.hub.Join <- &chat.Subscribe{
			UID:     uid,
			Channel: channel,
		}
		// logger.Info("subscribe channel succ", zap.String("uid", fmt.Sprintf("%d", uid)), zap.String("channel", channel))
	}
	return nil
}

func (s *ChatService) LeaveChannel(client *chat.Client, channel uint) {

}

func (s *ChatService) GetOfflineMessages(uid uint) []string {
	// 获取并发送离线消息
	return s.Repo.GetOfflineMessages(uid)
}

func (s *ChatService) GetChannelList() ([]*entities.ChatChannel, error) {
	return s.Repo.GetChannelList()
}

func (s *ChatService) GetMessageHistory(req *entities.GetMessageHistoryReq) error {
	return s.Repo.GetMessageHistory(req)
}
