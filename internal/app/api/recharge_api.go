package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/pay"
	"rk-api/internal/app/service"
	"rk-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var RechargeAPISet = wire.NewSet(wire.Struct(new(RechargeAPI), "*"))

type RechargeAPI struct {
	Srv *service.RechargeService
}

func (c *RechargeAPI) GetRechargeOrderList(ctx *gin.Context) {
	req := new(entities.GetRechargeOrderListReq)
	req.UID = ginx.Mine(ctx)
	req.Page = 1
	req.PageSize = 50
	err := c.Srv.GetRechargeOrderList(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *RechargeAPI) GetRechargeConfig(ctx *gin.Context) {
	info, err := c.Srv.GetRechargeConfig()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, info)
}

func (c *RechargeAPI) GetRechargeUrlInfo(ctx *gin.Context) {

	var req entities.GetRechargeUrlReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	info, err := c.Srv.GetRechargeUrlInfo(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, info)
}

func (c *RechargeAPI) KBCallback(ctx *gin.Context) {

	var back pay.KBBack
	if err := ctx.ShouldBind(&back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("RechargeAPI KBCallback", zap.Any("req", back))

	setting, err := c.Srv.GetRechargeChannelSetting("kb")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	if !back.VerifySignature(setting.PayKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.RechargeCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

type TKResponse struct {
	Code    int        `json:"code"`
	Data    pay.TKBack `json:"data"`
	Message string     `json:"message"`
}

func (c *RechargeAPI) TKCallback(ctx *gin.Context) {

	var req TKResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	back := req.Data
	if req.Code == 200 && req.Message == "SUCCESS" {
		back.TradeStatus = 1
	}

	logger.ZInfo("RechargeAPI TKCallback", zap.Any("req", req))
	setting, err := c.Srv.GetRechargeChannelSetting("tk")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if !back.VerifySignature(setting.PayKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.RechargeCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

type ATResponse struct {
	Data       pay.ATBack `json:"resource"`
	NotifyTime int64      `json:"notifyTime"`
	EventType  string     `json:"eventType"`
	ID         string     `json:"id"`
}

func (c *RechargeAPI) ATCallback(ctx *gin.Context) {
	var req ATResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("RechargeAPI ATCallback", zap.Any("req", req))

	back := req.Data
	signatureReceived := ctx.Request.Header.Get("X-Qu-Signature")
	if signatureReceived == "" {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}
	logger.ZInfo("RechargeAPI ATCallback", zap.String("X-Qu-Signature", signatureReceived))

	setting, err := c.Srv.GetRechargeChannelSetting("at")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	params := map[string]string{
		// "X-Qu-Signature-Version": "v1.0",
		"X-Qu-Signature-Version": ctx.Request.Header.Get("X-Qu-Signature-Version"),
		"X-Qu-Access-Key":        ctx.Request.Header.Get("X-Qu-Access-Key"),
		"X-Qu-Mid":               ctx.Request.Header.Get("X-Qu-Mid"),
		"X-Qu-Nonce":             ctx.Request.Header.Get("X-Qu-Nonce"),
		"X-Qu-Signature-Method":  ctx.Request.Header.Get("X-Qu-Signature-Method"),
		"X-Qu-Timestamp":         ctx.Request.Header.Get("X-Qu-Timestamp"),
	}

	logger.ZInfo("RechargeAPI ATCallback", zap.Any("header", params))

	if !back.VerifySignature(params, setting.PaySecret, signatureReceived) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.RechargeCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

type GOResponse struct {
	Code    int        `json:"code"`
	Data    pay.GOBack `json:"data"`
	Message string     `json:"message"`
}

func (c *RechargeAPI) GOCallback(ctx *gin.Context) {

	var req GOResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	back := req.Data
	if req.Code == 200 && req.Message == "SUCCESS" {
		back.TradeStatus = 1
	}

	logger.ZInfo("RechargeAPI GOCallback", zap.Any("req", req))
	setting, err := c.Srv.GetRechargeChannelSetting("go")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if !back.VerifySignature(setting.PayKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.RechargeCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

// CallbackRequest是异步回调接收的数据格式
type COWResponse struct {
	Transdata string `json:"transdata"`
	Sign      string `json:"sign"`
}

func (c *RechargeAPI) COWCallback(ctx *gin.Context) {

	var req COWResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("RechargeAPI COWCallback", zap.Any("req", req))
	// URL解码transdata
	transdata, err := url.QueryUnescape(req.Transdata)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	// // 解码transdata为JSON对象
	var back pay.COWBack
	if err := json.Unmarshal([]byte(transdata), &back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	back.Sign = req.Sign

	logger.ZInfo("RechargeAPI COWCallback", zap.Any("back", back))

	setting, err := c.Srv.GetRechargeChannelSetting("cow")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if !back.VerifySignature(setting.PayKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	_, err = c.Srv.RechargeCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	// 响应支付平台以确认收到通知
	ctx.Status(http.StatusOK)
}

// CallbackRequest是异步回调接收的数据格式
type ANTResponse struct {
	Transdata string `json:"transdata"`
	Sign      string `json:"sign"`
}

func (c *RechargeAPI) ANTCallback(ctx *gin.Context) {

	var req ANTResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("RechargeAPI ANTCallback", zap.Any("req", req))
	// URL解码transdata
	transdata, err := url.QueryUnescape(req.Transdata)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	// // 解码transdata为JSON对象
	var back pay.ANTBack
	if err := json.Unmarshal([]byte(transdata), &back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	back.Sign = req.Sign

	logger.ZInfo("RechargeAPI ANTCallback", zap.Any("back", back))

	setting, err := c.Srv.GetRechargeChannelSetting("ANT")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if !back.VerifySignature(setting.PayKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	_, err = c.Srv.RechargeCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	// 响应支付平台以确认收到通知
	ctx.Status(http.StatusOK)
}

func (c *RechargeAPI) DYCallback(ctx *gin.Context) {

	var back pay.DYBack
	if err := ctx.ShouldBind(&back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("RechargeAPI DYCallback", zap.Any("req", back))

	setting, err := c.Srv.GetRechargeChannelSetting("dy")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	if !back.VerifySignature(setting.PayKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.RechargeCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

func (c *RechargeAPI) GAGACallback(ctx *gin.Context) {

	var back pay.GaGaBack
	if err := ctx.ShouldBind(&back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("RechargeAPI GAGACallback", zap.Any("req", back))

	setting, err := c.Srv.GetRechargeChannelSetting("gaga")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	if !back.VerifySignature(setting.PayKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.RechargeCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

func (c *RechargeAPI) POLYCallback(ctx *gin.Context) {

	// var req pay.TKPay
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	ginx.RespErr(ctx, err)
	// 	return
	// }
	// setting, err := c.Srv.GetRechargeChannelSetting("poly")
	// if err != nil {
	// 	ctx.String(http.StatusOK, err.Error())
	// 	return
	// }

	// if !req.VerifySignature(setting.PayKey) {
	// 	ctx.String(http.StatusOK, "verify signature fail")
	// 	return
	// }

	// resp, err := c.Srv.RechargeCallbackProcess(req)
	// if err != nil {
	// 	ctx.String(http.StatusOK, err.Error())
	// 	return
	// }
	// ctx.String(http.StatusOK, resp)
}
