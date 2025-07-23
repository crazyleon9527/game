package api

import (
	"net/http"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var UserAPISet = wire.NewSet(wire.Struct(new(UserAPI), "*"))

type UserAPI struct {
	Srv *service.UserService
}

// 获取用户信息时检验登录
func (c *UserAPI) GetUserInfo(ctx *gin.Context) {

	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
		return
	}
	token := authHeader[7:] // Remove "Bearer " prefix
	uid := ginx.Mine(ctx)
	if err := c.Srv.VerifyLoginToken(uid, token); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	user, err := c.Srv.GetUserProfile(uid)
	// user, err := c.Srv.GetUserProfile(uid)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	now := time.Now()
	loginDate := time.Unix(user.LoginTime, 0)
	if now.Year() != loginDate.Year() ||
		now.YearDay() != loginDate.YearDay() {
		user.LoginTime = now.Unix()
		c.Srv.UpdateLoginInfo(user.ID, ctx.ClientIP())
	}
	ginx.RespSucc(ctx, user)
}

func (c *UserAPI) GetCustomer(ctx *gin.Context) {
	customer, err := c.Srv.GetCustomer(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, customer)
}

func (c *UserAPI) SearchUser(ctx *gin.Context) {

	var req entities.SearchUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	user, err := c.Srv.SearchUser(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, user)
}

func (c *UserAPI) GetBetUserInfo(ctx *gin.Context) {

	user, err := c.Srv.GetUserByUID(ginx.Mine(ctx))
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	betUser := entities.BetUser{
		UID: user.ID,
		// BetAmountLimit: user.BetAmountLimit,
		// BetTimesLimit:  user.BetTimesLimit,
		// UntilCash:      user.UntilCash,
		// UntilTime:      user.UntilTime,
	}
	ginx.RespSucc(ctx, betUser)
}

func (c *UserAPI) EditNickname(ctx *gin.Context) {
	var req entities.EditNicknameReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.ChangeNickname(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *UserAPI) EditAvatar(ctx *gin.Context) {
	var req entities.EditAvatarReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.ChangeAvatar(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *UserAPI) BindTelegram(ctx *gin.Context) {
	var req entities.BindTelegramReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.BindTelegram(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *UserAPI) BindEmail(ctx *gin.Context) {
	var req entities.BindEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.UID = ginx.Mine(ctx)
	err := c.Srv.BindEmail(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *UserAPI) EditUserInfo(ctx *gin.Context) {
	var req entities.EditUserInfoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.OptionID = ginx.Mine(ctx)
	req.IP = ctx.ClientIP()
	err := c.Srv.EditUserInfo(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *UserAPI) ClearUserCache(ctx *gin.Context) {
	var req entities.ClearUserCacheReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	err := c.Srv.ClearUserCache(req.UID)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *UserAPI) UploadAvatar(ctx *gin.Context) {
	// 获取上传的文件
	file, err := ctx.FormFile("avatar")
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	var uid = ginx.Mine(ctx)
	err = c.Srv.UpdateAvatar(uid, file)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}
