package api

import (
	"errors"
	"net/http"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/game/crash"
	"rk-api/internal/app/ginx"
	"rk-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var CrashGameAPISet = wire.NewSet(wire.Struct(new(CrashGameAPI), "*"))

type CrashGameAPI struct {
	CrashGame *crash.CrashGame
}

var crashupgrader = websocket.Upgrader{
	EnableCompression: false, // 禁用压缩以避免潜在的消息合并
	WriteBufferSize:   1024,  // 设置适当的写缓冲区大小
	ReadBufferSize:    1024,  // 设置适当的读缓冲区大小

	CheckOrigin: func(r *http.Request) bool { return true },
}

func (c *CrashGameAPI) WsHandler(ctx *gin.Context) {
	uid := ginx.Mine(ctx)
	logger.ZInfo("CrashGameAPI WsHandler", zap.Uint("uid", uid))
	conn, err := crashupgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}
	logger.ZInfo("CrashGameAPI upgrader Upgrade")

	client := c.CrashGame.Srv.Connect(uid, conn)
	logger.ZInfo("CrashGameAPI client connected")

	// 启动客户端
	client.Start()

	c.CrashGame.Srv.JoinChannel(uid, "all")
	logger.ZInfo("CrashGameAPI subscribe channel")
}

// GetCrashGameRound
func (c *CrashGameAPI) GetCrashGameRound(ctx *gin.Context) {
	var req *entities.GetCrashGameRoundReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	var rsp *entities.GetCrashGameRoundRsp
	if req.RoundID == 0 {
		rsp, err = c.CrashGame.GetCrashGameRound()
	} else {
		rsp, err = c.CrashGame.GetDBCrashGameRound(req.RoundID)
	}
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// GetCrashGameRoundList
func (c *CrashGameAPI) GetCrashGameRoundList(ctx *gin.Context) {
	rsp, err := c.CrashGame.GetCrashGameRoundList()
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// GetCrashGameRoundOrderList
func (c *CrashGameAPI) GetCrashGameRoundOrderList(ctx *gin.Context) {
	var req *entities.GetCrashGameRoundOrderListReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	var rsp []*entities.CrashGameOrder
	if req.RoundID == 0 {
		rsp, err = c.CrashGame.GetCrashGameRoundOrderList()
	} else {
		rsp, err = c.CrashGame.GetDBCrashGameRoundOrderList(req.RoundID)
	}
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// PlaceCrashGameBet
func (c *CrashGameAPI) PlaceCrashGameBet(ctx *gin.Context) {
	var req *entities.PlaceCrashGameBetReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := c.CrashGame.PlaceCrashGameBet(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// CancelCrashGameBet
func (c *CrashGameAPI) CancelCrashGameBet(ctx *gin.Context) {
	var req *entities.CancelCrashGameBetReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	err = c.CrashGame.CancelCrashGameBet(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

// EscapeCrashGameBet
func (c *CrashGameAPI) EscapeCrashGameBet(ctx *gin.Context) {
	var req *entities.EscapeCrashGameBetReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := c.CrashGame.EscapeCrashGameBet(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// GetCrashAutoBet
func (c *CrashGameAPI) GetCrashAutoBet(ctx *gin.Context) {
	uid := ginx.Mine(ctx)

	rsp, err := c.CrashGame.GetCrashAutoBet(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// PlaceCrashAutoBet
func (c *CrashGameAPI) PlaceCrashAutoBet(ctx *gin.Context) {
	var req *entities.PlaceCrashAutoBetReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	rsp, err := c.CrashGame.PlaceCrashAutoBet(req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// CancelCrashAutoBet
func (c *CrashGameAPI) CancelCrashAutoBet(ctx *gin.Context) {
	uid := ginx.Mine(ctx)

	err := c.CrashGame.CancelCrashAutoBet(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

// GetUserCrashGameOrder
func (c *CrashGameAPI) GetUserCrashGameOrder(ctx *gin.Context) {
	var req *entities.GetUserCrashGameOrderReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)

	var rsp []*entities.CrashGameOrder
	if req.RoundID == 0 {
		rsp, err = c.CrashGame.GetUserCrashGameOrder(req.UID)
	} else {
		rsp, err = c.CrashGame.GetDBUserCrashGameOrder(req.UID, req.RoundID)
	}
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// GetUserCrashGameOrderList
func (c *CrashGameAPI) GetUserCrashGameOrderList(ctx *gin.Context) {
	uid := ginx.Mine(ctx)

	rsp, err := c.CrashGame.GetUserCrashGameOrderList(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, rsp)
}

// Test
func (c *CrashGameAPI) Test(ctx *gin.Context) {
	seed := ctx.DefaultQuery("seed", "233")
	if seed == "" {
		ginx.RespErr(ctx, errors.New("seed is empty"))
		return
	}
	c.CrashGame.Test(seed)
	ginx.RespSucc(ctx, seed)
}
