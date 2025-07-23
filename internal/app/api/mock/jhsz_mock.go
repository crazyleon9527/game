package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// JhszSet 注入Jhsz
var JhszSet = wire.NewSet(wire.Struct(new(Jhsz), "*"))

type Jhsz struct {
}

// @Tags Jhsz
// @Summary 登录启动游戏
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.JhszGameLoginReq true "params"
// @Success 200 {object}  map[string]interface{}
// @Router /api/jhszgame/launch [post]
func (a *Jhsz) Launch(c *gin.Context) {
}

// @Tags Jhsz
// @Summary 发送通知(内部用,前端不需要接入)
// @Accept  json
// @Produce  json
// @Param req body entities.SendNotificationReq true "params"
// @Success 200 {object}  map[string]interface{}
// @Router /api/jhszgame/send-notification [post]
func (a *Jhsz) SendNotification(c *gin.Context) {
}
