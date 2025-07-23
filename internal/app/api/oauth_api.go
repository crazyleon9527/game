package api

import (
	"fmt"
	"net/http"
	"rk-api/internal/app/config"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/ginx"
	"rk-api/internal/app/service"
	"rk-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var OauthAPISet = wire.NewSet(wire.Struct(new(OauthAPI), "*"))

type OauthAPI struct {
	Srv     *service.OauthService
	UserSrv *service.UserService
	AuthSrv *service.AuthService
}

type OAuthLoginReq struct {
	Plat        string `json:"plat" example:"github" binding:"required"`                         //第三方平台(google,facebook,twister,github[可用])
	RedirectUrl string `json:"redirect_url" example:"https://www.baidu.com" binding:"omitempty"` //第三方登录成功后跳转地址
	Extra       string `json:"extra" example:"" binding:"omitempty"`                             //第三方登录成功后跳转地址携带参数
}

func (c *OauthAPI) Login(ctx *gin.Context) {
	var req entities.OAuthLoginReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	oauthState := entities.OauthState{
		Extra:       req.Extra,
		RedirectUrl: req.RedirectUrl,
		Plat:        req.Plat,
	}

	logger.ZInfo("Login oauth state", zap.Any("oauth state", oauthState))

	if err := c.Srv.OauthStateUrl(&oauthState); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	// ginx.RespSucc(ctx, oauthState)
	// ctx.Redirect(http.StatusTemporaryRedirect, oauthState.OauthUrl)
	ginx.RespSucc(ctx, oauthState)
	// c.Redirect(http.StatusMovedPermanently, oauthState.OauthUrl)

}

func (c *OauthAPI) GoogleCallBack(ctx *gin.Context) {

	state, _ := ctx.GetQuery("state")
	code, _ := ctx.GetQuery("code")
	oauthState, oerr := c.Srv.GetOauthState(state)
	if oerr != nil {
		ginx.RespErr(ctx, oerr)
		return
	}

	gu, err := c.Srv.VerifyGoogleUser(code)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	email := gu.Email
	// 判断是否存在
	user, err := c.UserSrv.GetUserByname(email)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	if user == nil {
		// 创建用户
		reg := entities.RegisterCredentials{
			Username: gu.Email,
			Email:    gu.Email,
			VerCode:  config.Get().ServiceSettings.TrustedUserCode,
			IsOAuth:  true,
		}

		user, err = c.AuthSrv.RegisterUser(&reg)
		if err != nil {
			ginx.RespErr(ctx, err)
			return
		}
	}

	ginx.RespSucc(ctx, oauthState)
	// c.Redirect(http.StatusMovedPermanently, oauthState.OauthUrl)
	sso := entities.OAuthToken{
		UID: user.ID,
	}

	err = c.AuthSrv.CreateJWTAccessRefreshToken(&sso)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.SetAuthTokens(ctx, sso.AccessToken, sso.RefreshToken, config.Get().ServiceSettings.TokenRefreshTime)
	ctx.Redirect(http.StatusMovedPermanently, oauthState.GetUrl())
}

func (c *OauthAPI) CallBackGithub(ctx *gin.Context) {
	state, _ := ctx.GetQuery("state")
	code, _ := ctx.GetQuery("code")
	oauthState, oerr := c.Srv.GetOauthState(state)
	if oerr != nil {
		ginx.RespErr(ctx, oerr)
		return
	}

	oauthUser, err := c.Srv.VerifyGithubUser(code)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	oauthID := fmt.Sprintf("github_%d", *oauthUser.ID)

	// 判断是否存在
	user, err := c.UserSrv.GetUserByname(oauthID)

	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}

	if user == nil {
		email := ""
		if oauthUser.Email != nil {
			email = *oauthUser.Email
		}
		// 创建用户
		reg := entities.RegisterCredentials{
			Username: oauthID,
			Email:    email,
			VerCode:  config.Get().ServiceSettings.TrustedUserCode,
			IsOAuth:  true,
		}

		user, err = c.AuthSrv.RegisterUser(&reg)
		if err != nil {
			ginx.RespErr(ctx, err)
			return
		}
	}

	// c.Redirect(http.StatusMovedPermanently, oauthState.OauthUrl)
	sso := entities.OAuthToken{
		UID: user.ID,
	}

	err = c.AuthSrv.CreateJWTAccessRefreshToken(&sso)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.SetAuthTokens(ctx, sso.AccessToken, sso.RefreshToken, config.Get().ServiceSettings.TokenRefreshTime)

	ctx.Redirect(http.StatusMovedPermanently, oauthState.GetUrl())
}

func (c *AuthAPI) ReigsterAuthUser(ctx *gin.Context) {
	var req entities.AddUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	req.OptionID = ginx.Mine(ctx)
	req.IP = ctx.ClientIP()
	err := c.Srv.ReigsterAuthUser(&req)
	if err != nil {
		ginx.RespErr(ctx, err)
		return
	}
	ginx.RespSucc(ctx, nil)
}
