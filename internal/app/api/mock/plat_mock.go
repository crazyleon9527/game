package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// PlatSet 注入Plat
var PlatSet = wire.NewSet(wire.Struct(new(Plat), "*"))

type Plat struct {
}

// @Tags Plat
// @Summary 获取平台配置信息
// @Description 获取平台配置信息,汇率点等
// @Accept  json
// @Produce  json
// @Success 200 {object} entities.PlatSetting
// @Router /api/plat/get-platform [post]
func (a *Plat) GetPlatSetting(c *gin.Context) {
}
