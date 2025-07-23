package service

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/mq"
	"rk-api/internal/app/mq/handle"
	"rk-api/internal/app/service/repository"
	"rk-api/pkg/logger"

	"github.com/google/wire"
	"go.uber.org/zap"
)

var NotificationServiceSet = wire.NewSet(
	ProvideNotificationService,
)

type NotificationService struct {
	Repo    *repository.NotificationRepository
	UserSrv *UserService
}

func ProvideNotificationService(
	repo *repository.NotificationRepository,
	userSrv *UserService,

) *NotificationService {
	service := &NotificationService{
		Repo:    repo,
		UserSrv: userSrv,
	}
	return service
}
func (s *NotificationService) CreateNotification(userID uint, message string) error {
	notification := &entities.Notification{
		UID:     userID,
		Message: message,
		Read:    0,
	}
	return s.Repo.CreateNotification(notification)
}

func (s *NotificationService) GetNotificationList(req *entities.GetNotificationListReq) error {
	return s.Repo.GetNotificationList(req)
}

func (s *NotificationService) MarkAsRead(notificationID uint) error {
	return s.Repo.MarkNotificationAsRead(notificationID)
}

func (s *NotificationService) MarkAllAsRead(uid uint) error {
	return s.Repo.MarkAllAsRead(uid)
}
func (s *NotificationService) GetUnreadNotificationCount(uid uint) (int64, error) {
	return s.Repo.GetUnreadNotificationCount(uid)
}

func (m *NotificationService) HandleNotification(notification *entities.Notification) error { //处理通知
	// logger.ZInfo("HandleNotification", zap.Any("notification", notification))
	return m.Repo.CreateNotification(notification)
}

func SendNotification(uid uint, message string, title string) { //发送通知
	notificationQueue, _ := handle.NewNotificationQueue(&entities.Notification{UID: uid, Message: message, Title: title})
	if _, err := mq.MClient.Enqueue(notificationQueue); err != nil {
		logger.ZInfo("notificationQueue fail", zap.Error(err))
	}
}

func (s *NotificationService) SendNotification(req *entities.SendNotificationReq) {
	notification := entities.Notification{
		UID:        req.UID,
		Type:       req.Type,
		TemplateID: req.TemplateID,
		Params:     req.Params,
		Message:    req.Message,
		Read:       0,
	}
	notificationQueue, _ := handle.NewNotificationQueue(&notification)
	if _, err := mq.MClient.Enqueue(notificationQueue); err != nil {
		logger.ZInfo("notificationQueue fail", zap.Error(err))
	}

}
