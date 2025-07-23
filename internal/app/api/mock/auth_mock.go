package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// AuthSet 注入Auth
var AuthSet = wire.NewSet(wire.Struct(new(Auth), "*"))

// Auth 示例程序
type Auth struct {
}

// @Tags Auth
// @Summary 注册用户
// @Accept  json
// @Produce  json
// @Param req body entities.RegisterCredentials true "params"
// @Success 200
// @Router /api/auth/register [post]
func (c *Auth) RegisterUser(ctx *gin.Context) {
}

// @Tags Auth
// @Summary 登录
// @Accept  json
// @Produce  json
// @Param req body entities.MobileLoginCredentials true "params"
// @Success 200
// @Router /api/auth/mobile-login [post]
func (c *Auth) MobileLogin(ctx *gin.Context) {
}

// @Tags Auth
// @Summary 登录
// @Accept  json
// @Produce  json
// @Param req body entities.LoginCredentials true "params"
// @Success 200
// @Router /api/auth/login [post]
func (c *Auth) Login(ctx *gin.Context) {
}

// @Tags Auth
// @Summary 退出登录
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200
// @Router /api/auth/logout [post]
func (c *Auth) Logout(ctx *gin.Context) {
}

// @Tags Auth
// @Summary 更改密码
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.ChangePasswordCredentials true "params"
// @Success 200 {object} ginx.Resp
// @Router /api/auth/update-password [post]
func (c *Auth) ChangePassword(ctx *gin.Context) {
}

// @Tags Auth
// @Summary 重置密码
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.ResetPasswordCredentials true "params"
// @Success 200 {object} ginx.Resp
// @Router /api/auth/reset-password [post]
func (c *Auth) ResetPassword(ctx *gin.Context) {
}
