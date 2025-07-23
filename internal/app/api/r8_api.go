package api

import (
	"net/http"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var R8APISet = wire.NewSet(wire.Struct(new(R8API), "*"))

type R8API struct {
	Srv *service.R8Service
}

func (c *R8API) Login(ctx *gin.Context) {
	var req entities.R8GameLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	data, err := c.Srv.Login(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *R8API) Kick(ctx *gin.Context) {
	data, err := c.Srv.Kick(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, data)
}

func (c *R8API) GetSessionIDToken(ctx *gin.Context) {
	apiKey := ctx.GetHeader("api-key")
	pfid := ctx.GetHeader("pf-id")
	timestamp := ctx.GetHeader("timestamp")
	data := c.Srv.GetSessionIDToken(apiKey, pfid, timestamp)
	ctx.JSON(http.StatusOK, data)
}

// 提取Authorization头
// authorization := c.GetHeader("Authorization")
// if authorization == "" {
//     c.JSON(http.StatusUnauthorized, gin.H{
//         "code": 22001,
//         "msg":  "Authorization header missing",
//     })
//     return
// }

func (c *R8API) GetBalance(ctx *gin.Context) {
	account := ctx.Param("uid")
	data := c.Srv.FetchBalance(account)
	ctx.JSON(http.StatusOK, data)
}

func (c *R8API) Transfer(ctx *gin.Context) {
	var req entities.R8Transfer
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 22007, "msg": "Bad request"})
		return
	}
	data := c.Srv.Transfer(&req)
	ctx.JSON(http.StatusOK, data)
}

func (c *R8API) AwardActivity(ctx *gin.Context) {
	var req entities.R8Activity
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 22007, "msg": "Bad request"})
		return
	}
	data := c.Srv.AwardActivity(&req)
	ctx.JSON(http.StatusOK, data)
}
