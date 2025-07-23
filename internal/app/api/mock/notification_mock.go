package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// NotificationSet 注入Notification
var NotificationSet = wire.NewSet(wire.Struct(new(Notification), "*"))

type Notification struct {
}

// @Summary 获取通知列表
// @Description 根据用户ID获取通知列表，支持分页查询
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.GetNotificationListReq true "请求参数"
// @Success 200 {object} entities.Paginator "返回分页数据"
// @Router /api/notification/get-notification-list [post]
func (c *Notification) GetNotificationList(ctx *gin.Context) {

}

// MarkAsRead 标记通知为已读
// @Summary 标记通知为已读
// @Description 根据通知ID标记指定通知为已读
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body entities.IDReq true "通知ID"
// @Success 200
// @Router /api/notification/mark-as-read [post]
func (c *Notification) MarkAsRead(ctx *gin.Context) {

}

// MarkAllAsRead 标记所有通知为已读
// @Summary 标记所有未读通知为已读
// @Description 根据用户ID标记所有未读通知为已读
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Router /api/notification/mark-all-as-read [post]
func (c *Notification) MarkAllAsRead(ctx *gin.Context) {

}

// GetUnreadNotificationCount 获取未读通知数量
// @Summary 获取未读通知数量
// @Description 获取用户的未读通知数量
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200
// @Router /api/notification/get-unread-notification-count [post]
func (c *Notification) GetUnreadNotificationCount(ctx *gin.Context) {

}
