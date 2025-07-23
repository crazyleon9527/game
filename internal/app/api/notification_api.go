package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var NotificationAPISet = wire.NewSet(wire.Struct(new(NotificationAPI), "*"))

type NotificationAPI struct {
	Srv *service.NotificationService
}

func (c *NotificationAPI) GetNotificationList(ctx *gin.Context) {
	var req entities.GetNotificationListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	err := c.Srv.GetNotificationList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *NotificationAPI) MarkAsRead(ctx *gin.Context) {
	var req entities.IDReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.MarkAsRead(req.ID)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *NotificationAPI) MarkAllAsRead(ctx *gin.Context) {
	err := c.Srv.MarkAllAsRead(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *NotificationAPI) GetUnreadNotificationCount(ctx *gin.Context) {
	count, err := c.Srv.GetUnreadNotificationCount(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, count)
}
