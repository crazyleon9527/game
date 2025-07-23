package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var FlowAPISet = wire.NewSet(wire.Struct(new(FlowAPI), "*"))

type FlowAPI struct {
	Srv *service.FlowService
}

func (c *FlowAPI) GetFlowList(ctx *gin.Context) {
	var req entities.GetFlowListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	if req.UID == 0 {
		req.UID = ginx.Mine(ctx)
	}

	err := c.Srv.GetFlowList(&req)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}
