package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// @BasePath
// @Summary 第三方登录
// @Param   OAuthReq body schema.OAuthLoginReq true "第三方登录"
// @Tags    oauth
// @Router  /oauth/login [post]

// OauthSet 注入Oauth
var OauthSet = wire.NewSet(wire.Struct(new(Oauth), "*"))

// Oauth 示例程序
type Oauth struct {
}

// @Tags Oauth
// @Summary 第三方登录
// @Accept  json
// @Produce  json
// @Param req body entities.OAuthLoginReq true "第三方登录"
// @Success 200
// @Router /api/oauth/login [post]
func (a *Oauth) Login(c *gin.Context) {
}

// @Summary google登录回调
// @Tags    oauth
// @Router  /oauth/google/callback [get]
func (a *Oauth) GoogleCallBack(c *gin.Context) {
}

// @Summary github登录回调
// @Tags    oauth
// @Router  /oauth/github/callback [get]
func (a *Oauth) GithubCallBack(c *gin.Context) {
}
