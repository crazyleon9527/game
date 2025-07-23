package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var TransactionAPISet = wire.NewSet(wire.Struct(new(TransactionAPI), "*"))

type TransactionAPI struct {
	Srv *service.TransactionService
}

func (c *TransactionAPI) GetTransactionList(ctx *gin.Context) {
	var req entities.GetTransactionListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.GetTransactionList(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}
