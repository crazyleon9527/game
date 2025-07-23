package repository

import (
	"bytes"
	"encoding/json"
	"rk-api/internal/app/entities"
	"rk-api/pkg/logger"
	"text/template"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var NotificationRepositorySet = wire.NewSet(wire.Struct(new(NotificationRepository), "*"))

type NotificationRepository struct {
	DB *gorm.DB
}

func (r *NotificationRepository) CreateNotification(notification *entities.Notification) error {
	return r.DB.Create(notification).Error
}

func (r *NotificationRepository) GetNotificationList(param *entities.GetNotificationListReq) error {
	type notificationWithTemplate struct {
		entities.Notification
		TemplateContent *string // 模板内容可能为空
	}
	param.List = make([]*notificationWithTemplate, 0)

	logger.ZInfo("GetNotificationList", zap.Any("param", param))

	tx := r.DB.Table("notification").
		Select("notification.*,notification.created_at as created_at, notification_template.content as template_content").
		Joins("LEFT JOIN notification_template ON notification.template_id = notification_template.id").
		Where("notification.uid = ?", param.UID).
		Order("notification.created_at desc")

	// 分页
	if err := param.Paginate(tx); err != nil {
		return err
	}

	var notifications = param.List.([]*notificationWithTemplate)
	// 组装数据并渲染模板
	var list = make([]*entities.Notification, 0, len(notifications))
	for _, n := range notifications {
		// 如果是模板消息，渲染模板
		if n.TemplateID != nil && n.TemplateContent != nil {
			if n.Params == nil { // 如果没有参数，直接返回模板
				n.Message = *n.TemplateContent
			} else {
				n.Message = renderTemplate(*n.TemplateContent, *n.Params) //有参数则渲染模板
			}
		}
		list = append(list, &n.Notification)
	}
	param.List = list
	return nil
}

func renderTemplate(content string, params string) string {
	var data map[string]string
	if err := json.Unmarshal([]byte(params), &data); err != nil {
		return content // 如果解析失败，返回原始模板
	}

	tmpl, err := template.New("notification").Parse(content)
	if err != nil {
		return content
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return content
	}
	return buf.String()
}

func (r *NotificationRepository) MarkNotificationAsRead(notificationID uint) error {
	return r.DB.Model(&entities.Notification{}).Where("id = ?", notificationID).Update("read", 1).Error
}

func (r *NotificationRepository) GetUnreadNotificationCount(userID uint) (int64, error) {
	var count int64
	err := r.DB.Model(&entities.Notification{}).Where("uid = ? AND `read` = ?", userID, 0).Count(&count).Error
	return count, err
}

func (r *NotificationRepository) MarkAllAsRead(userID uint) error {
	return r.DB.Model(&entities.Notification{}).Where("uid = ?", userID).Update("read", 1).Error
}
