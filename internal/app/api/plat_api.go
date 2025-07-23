package api

import (
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var PlatAPISet = wire.NewSet(wire.Struct(new(PlatAPI), "*"))

type PlatAPI struct {
	Srv *service.PlatService
}

func (c *PlatAPI) GetPlatformInfo(ctx *gin.Context) {

	setting, err := c.Srv.GetPlatSetting()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, setting)
}
