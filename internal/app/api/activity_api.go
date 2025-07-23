package api

import (
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var ActivityAPISet = wire.NewSet(wire.Struct(new(ActivityAPI), "*"))

type ActivityAPI struct {
	Srv *service.ActivityService
}

func (c *ActivityAPI) AddRedEnvelope(ctx *gin.Context) {
	var req entities.AddRedEnvelopeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	req.IP = ctx.ClientIP()

	err := c.Srv.AddRedEnvelope(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, nil)
}

func (c *ActivityAPI) DelRedEnvelope(ctx *gin.Context) {
	var req entities.DelRedEnvelopeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.IP = ctx.ClientIP()

	err := c.Srv.DelRedEnvelope(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *ActivityAPI) GetActivityList(ctx *gin.Context) {

	list, err := c.Srv.GetActivityList()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, list)
}

func (c *ActivityAPI) GetBannerList(ctx *gin.Context) {
	list, err := c.Srv.GetBannerList()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, list)
}

// GetLogoList
func (c *ActivityAPI) GetLogoList(ctx *gin.Context) {
	logoType := ctx.DefaultQuery("logoType", "1")
	// 转换为整数
	ltype, err := strconv.Atoi(logoType)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	list, err := c.Srv.GetLogoList(ltype)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, list)
}

func (c *ActivityAPI) GetRedEnvelope(ctx *gin.Context) {
	var req entities.GetRedEnvelopeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	if req.Type == 0 {
		amount, err := c.Srv.GetRedEnvelopeAmount(req.RedName)
		if err != nil {
			ginx.RespErr(ctx, err)
			return
		}
		ginx.RespSucc(ctx, gin.H{"amount": amount})
		return
	}

	err := c.Srv.GetRedEnvelope(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, nil)
}

func (c *ActivityAPI) JoinPinduo(ctx *gin.Context) {
	var req entities.GetPinduoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	pd, err := c.Srv.JoinPinDuo(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, pd)
}

func (c *ActivityAPI) GetPinduoCash(ctx *gin.Context) {
	var req entities.GetPinduoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	pd, err := c.Srv.ReceivePinduoBonus(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, pd)
}
