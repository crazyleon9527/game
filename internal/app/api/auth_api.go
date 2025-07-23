package api

import (
	"net/http"
	"rk-api/internal/app/config"
	"rk-api/internal/app/constant"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/errors"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"
	"rk-api/internal/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var AuthAPISet = wire.NewSet(wire.Struct(new(AuthAPI), "*"))

type AuthAPI struct {
	Srv *service.AuthService
}

func (c *AuthAPI) RegisterUser(ctx *gin.Context) {

	var credentials entities.RegisterCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	if len(credentials.Username) != 0 {
		if !utils.IsValidUsername(credentials.Username) {
			ginx.RespErr(ctx, errors.WithCode(errors.InvalidUsername))
			return
		}
		credentials.VerCode = config.Get().ServiceSettings.TrustedUserCode
	}

	userAgent := ctx.Request.Header.Get("User-Agent")
	credentials.Device = utils.GetDeviceOS(userAgent)

	credentials.LoginIP = ctx.ClientIP()

	_, err := c.Srv.RegisterUser(&credentials)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *AuthAPI) VerifyCredentials(ctx *gin.Context) {
	var credentials entities.VerifyCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Srv.VerifyCredentials(&credentials)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}

func (c *AuthAPI) MobileLogin(ctx *gin.Context) {
	var credentials entities.MobileLoginCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	credentials.LoginIP = ctx.ClientIP()

	user, err := c.Srv.MobileLogin(&credentials)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	// 在此处验证用户名和密码，并生成 JWT
	sso := entities.OAuthToken{
		UID: user.ID,
	}
	err = c.Srv.CreateJWTAccessRefreshToken(&sso)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.SetAuthTokens(ctx, sso.AccessToken, sso.RefreshToken, config.Get().ServiceSettings.TokenExpireTime)

	loginResp := &entities.LoginResp{
		Token: sso.AccessToken,
		UID:   user.ID,
	}
	ginx.RespSucc(ctx, loginResp)
}

func (c *AuthAPI) Login(ctx *gin.Context) {
	var credentials entities.LoginCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	credentials.LoginIP = ctx.ClientIP()

	user, err := c.Srv.Login(&credentials)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	// 在此处验证用户名和密码，并生成 JWT
	sso := entities.OAuthToken{
		UID: user.ID,
	}
	err = c.Srv.CreateJWTAccessRefreshToken(&sso)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.SetAuthTokens(ctx, sso.AccessToken, sso.RefreshToken, config.Get().ServiceSettings.TokenExpireTime)

	loginResp := &entities.LoginResp{
		Token: sso.AccessToken,
		UID:   user.ID,
	}

	ginx.RespSucc(ctx, loginResp)
}

func (c *AuthAPI) Logout(ctx *gin.Context) {

	userID := ginx.Mine(ctx)

	c.Srv.Logout(userID)

	if err := c.Srv.Logout(userID); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	secure := utils.IsHttps(ctx)
	ctx.SetCookie(constant.ACCESS_TOKEN, "", -1, "/", "", secure, true)
	ctx.SetCookie(constant.REFRESH_TOKEN, "", -1, "/", "", secure, true)
	ginx.RespSucc(ctx, nil)
}

func (c *AuthAPI) ChangePassword(ctx *gin.Context) {
	var credentials entities.ChangePasswordCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	if credentials.NewPassword != credentials.ConfirmPassword {
		ginx.RespErr(ctx, errors.With("new passwords do not match"))
		return
	}

	err := c.Srv.ChangePassword(&credentials)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, nil)
}

func (c *AuthAPI) ResetPassword(ctx *gin.Context) {
	var credentials entities.ResetPasswordCredentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	err := c.Srv.ResetPassword(&credentials)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	ginx.RespSucc(ctx, nil)
}

func (c *AuthAPI) Ping(ctx *gin.Context) {
	ctx.String(http.StatusOK, "pong")
}
