package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var GameAPISet = wire.NewSet(wire.Struct(new(GameAPI), "*"))

type GameAPI struct {
	Srv *service.GameService
}

func (c *GameAPI) GetGameList(ctx *gin.Context) {
	var req entities.GetGameListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Srv.GetGameList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *GameAPI) SearchGame(ctx *gin.Context) {
	var req entities.SearchGameReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.SearchGame(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *GameAPI) RefreshGameList(ctx *gin.Context) {
	list, err := c.Srv.GetRefreshGameList()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, list)
}
