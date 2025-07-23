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

var WithdrawAPISet = wire.NewSet(wire.Struct(new(WithdrawAPI), "*"))

type WithdrawAPI struct {
	Srv *service.WithdrawService
}

func (c *WithdrawAPI) GetWithdrawDetail(ctx *gin.Context) {

	detail, err := c.Srv.GetWithdrawDetail(ginx.Mine(ctx))

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, detail)
}

func (c *WithdrawAPI) GetithdrawCardList(ctx *gin.Context) {

	detail, err := c.Srv.GetWithdrawCardListByUID(ginx.Mine(ctx))

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, detail)
}

func (c *WithdrawAPI) GetHallWidthdrawRecordList(ctx *gin.Context) {
	req := new(entities.GetHallWithdrawRecordListReq)
	req.UID = ginx.Mine(ctx)
	req.Page = 1
	req.PageSize = 20
	err := c.Srv.GetHallWithdrawRecordList(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, &req.Paginator)
}

func (c *WithdrawAPI) ApplyForWithdrawal(ctx *gin.Context) {
	var req entities.ApplyForWithdrawalReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	err := c.Srv.ApplyForWithdrawal(&req)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WithdrawAPI) AddWithdrawCard(ctx *gin.Context) {
	var req entities.AddWithdrawCardReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	err := c.Srv.AddWithdrawCard(&req)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WithdrawAPI) DelWithdrawCard(ctx *gin.Context) {
	var req entities.IDReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.DelWithdrawCardByID(ginx.Mine(ctx), req.ID)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WithdrawAPI) SelectWithdrawCard(ctx *gin.Context) {
	var req entities.IDReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Srv.SelectWithdrawCard(ginx.Mine(ctx), req.ID)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WithdrawAPI) ReviewWithdrawal(ctx *gin.Context) {
	var req entities.ReviewWithdrawalReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.SysUID = ginx.Mine(ctx)
	req.IP = ctx.ClientIP()
	err := c.Srv.ReviewWithdrawal(&req)
	// logger.ZError("---------------------2------", zap.Any("req", req), zap.Error(err))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WithdrawAPI) AddUserWithdrawCard(ctx *gin.Context) {
	var req entities.AddUserWithdrawCardReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.IP = ctx.ClientIP()
	err := c.Srv.AddUserWithdrawCard(&req)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WithdrawAPI) DelUserWithdrawCard(ctx *gin.Context) {
	var req entities.DelUserWithdrawCardReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.IP = ctx.ClientIP()
	err := c.Srv.DelUserWithdrawCard(&req)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WithdrawAPI) FixUserWithdrawCard(ctx *gin.Context) {
	var req entities.FixUserWithdrawCardReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.IP = ctx.ClientIP()
	err := c.Srv.FixUserWithdrawCard(&req)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *WithdrawAPI) KBCallback(ctx *gin.Context) {

	var back pay.WithdrawKBBack
	if err := ctx.ShouldBind(&back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("WithdrawAPI KBCallback", zap.Any("req", back))

	setting, err := c.Srv.GetRechargeChannelSetting("kb")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	if !back.VerifySignature(setting.WithdrawKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.WithdrawCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

type TKWithdrawResponse struct {
	Code    int                `json:"code"`
	Data    pay.WithdrawTKBack `json:"data"`
	Message string             `json:"message"`
}

func (c *WithdrawAPI) TKCallback(ctx *gin.Context) {

	var req TKWithdrawResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("WithdrawAPI TKCallback", zap.Any("req", req))

	back := req.Data
	if req.Code == 200 && req.Message == "SUCCESS" {
		back.TradeStatus = 1
	}

	setting, err := c.Srv.GetRechargeChannelSetting("tk")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if !back.VerifySignature(setting.WithdrawKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}
	resp, err := c.Srv.WithdrawCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

type ATWithdrawResponse struct {
	Data       pay.WithdrawATBack `json:"resource"`
	NotifyTime int64              `json:"notifyTime"`
	EventType  string             `json:"eventType"`
	ID         string             `json:"id"`
}

func (c *WithdrawAPI) ATCallback(ctx *gin.Context) {
	var req ATWithdrawResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	back := req.Data

	logger.ZInfo("WithdrawAPI ATCallback", zap.Any("req", req))

	signatureReceived := ctx.Request.Header.Get("X-Qu-Signature")
	if signatureReceived == "" {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	logger.ZInfo("WithdrawAPI ATCallback", zap.String("X-Qu-Signature", signatureReceived))

	setting, err := c.Srv.GetRechargeChannelSetting("at")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	params := map[string]string{
		"X-Qu-Signature-Version": ctx.Request.Header.Get("X-Qu-Signature-Version"),
		"X-Qu-Access-Key":        ctx.Request.Header.Get("X-Qu-Access-Key"),
		"X-Qu-Mid":               ctx.Request.Header.Get("X-Qu-Mid"),
		"X-Qu-Nonce":             ctx.Request.Header.Get("X-Qu-Nonce"),
		"X-Qu-Signature-Method":  ctx.Request.Header.Get("X-Qu-Signature-Method"),
		"X-Qu-Timestamp":         ctx.Request.Header.Get("X-Qu-Timestamp"),
	}

	logger.ZInfo("WithdrawAPI ATCallback", zap.Any("header", params))

	if !back.VerifySignature(params, setting.PaySecret, signatureReceived) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.WithdrawCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

type GOWithdrawResponse struct {
	Code    int                `json:"code"`
	Data    pay.WithdrawGOBack `json:"data"`
	Message string             `json:"message"`
}

func (c *WithdrawAPI) GOCallback(ctx *gin.Context) {

	var req GOWithdrawResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("WithdrawAPI GOCallback", zap.Any("req", req))

	back := req.Data
	if req.Code == 200 && req.Message == "SUCCESS" {
		back.TradeStatus = 1
	}

	setting, err := c.Srv.GetRechargeChannelSetting("go")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if !back.VerifySignature(setting.WithdrawKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}
	resp, err := c.Srv.WithdrawCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

// CallbackRequest 是异步回调接收的完整数据结构
type COWWithdrawResponse struct {
	Transdata string `json:"transdata"`
	Sign      string `json:"sign"`
}

func (c *WithdrawAPI) COWCallback(ctx *gin.Context) {

	var req COWWithdrawResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	logger.ZInfo("WithdrawAPI COWCallback", zap.Any("req", req))
	transdata, err := url.QueryUnescape(req.Transdata)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	// // 解码transdata为JSON对象
	var back pay.WithdrawCOWBack
	if err := json.Unmarshal([]byte(transdata), &back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	back.Sign = req.Sign

	logger.ZInfo("WithdrawAPI COWCallback", zap.Any("back", back))

	setting, err := c.Srv.GetRechargeChannelSetting("cow")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if !back.VerifySignature(setting.PayKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	_, err = c.Srv.WithdrawCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	// 响应支付平台以确认收到通知
	ctx.Status(http.StatusOK)

}

type ANTWithdrawResponse struct {
	Transdata string `json:"transdata"`
	Sign      string `json:"sign"`
}

func (c *WithdrawAPI) ANTCallback(ctx *gin.Context) {

	var req ANTWithdrawResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	logger.ZInfo("WithdrawAPI ANTCallback", zap.Any("req", req))
	transdata, err := url.QueryUnescape(req.Transdata)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	// // 解码transdata为JSON对象
	var back pay.WithdrawANTBack
	if err := json.Unmarshal([]byte(transdata), &back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	back.Sign = req.Sign

	logger.ZInfo("WithdrawAPI ANTCallback", zap.Any("back", back))

	setting, err := c.Srv.GetRechargeChannelSetting("ANT")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	if !back.VerifySignature(setting.PayKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	_, err = c.Srv.WithdrawCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	// 响应支付平台以确认收到通知
	ctx.Status(http.StatusOK)

}

func (c *WithdrawAPI) DYCallback(ctx *gin.Context) {

	var back pay.WithdrawDYBack
	if err := ctx.ShouldBind(&back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("WithdrawAPI DYCallback", zap.Any("req", back))

	setting, err := c.Srv.GetRechargeChannelSetting("dy")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	if !back.VerifySignature(setting.WithdrawKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.WithdrawCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}

func (c *WithdrawAPI) GAGACallback(ctx *gin.Context) {

	var back pay.WithdrawGaGaBack
	if err := ctx.ShouldBind(&back); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	logger.ZInfo("WithdrawAPI GAGACallback", zap.Any("req", back))

	setting, err := c.Srv.GetRechargeChannelSetting("gaga")
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	if !back.VerifySignature(setting.WithdrawKey) {
		ctx.String(http.StatusOK, "verify signature fail")
		return
	}

	resp, err := c.Srv.WithdrawCallbackProcess(back)
	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	ctx.String(http.StatusOK, resp)
}
