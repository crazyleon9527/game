package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// UserSet 注入User
var UserSet = wire.NewSet(wire.Struct(new(User), "*"))

// User 示例程序
type User struct {
}

// @Tags User
// @Summary 获取用户信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entities.UserProfile
// @Router /api/user/get-user-info [post]
func (c *User) GetUserInfo(ctx *gin.Context) {
}

// @Tags User
// @Summary 编辑昵称
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.EditNicknameReq true "params"
// @Success 200 {object} ginx.Resp{}
// @Router /api/user/edit-nickname [post]
func (c *User) EditNickname(ctx *gin.Context) {
}

// @Tags User
// @Summary 编辑头像
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.EditAvatarReq true "params"
// @Success 200 {object} ginx.Resp{}
// @Router /api/user/edit-avatar [post]
func (c *User) EditAvatar(ctx *gin.Context) {
}

// @Tags User
// @Summary 绑定telegram
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.BindTelegramReq true "params"
// @Success 200 {object} ginx.Resp{}
// @Router /api/user/bind-telegram [post]
func (c *User) BindTelegram(ctx *gin.Context) {
}

// @Tags User
// @Summary 绑定email
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.BindEmailReq true "params"
// @Success 200 {object} ginx.Resp{}
// @Router /api/user/bind-email [post]
func (c *User) BindEmail(ctx *gin.Context) {
}

// @Tags User
// @Summary 上传头像
// @Accept  multipart/form-data
// @Produce  json
// @Security ApiKeyAuth
// @Param avatar formData file true "上传的头像文件"
// @Success 200 {object} ginx.Resp{}
// @Failure 400 {object} ginx.Resp{}
// @Failure 500 {object} ginx.Resp{}
// @Router /api/user/upload-avatar [post]
func (c *User) UploadAvatar(ctx *gin.Context) {
}

// @Tags User
// @Summary 抢红包
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.GetRedEnvelopeReq true "params"
// @Success 200 {object} ginx.Resp{}
// @Router /api/user/get-red-envelope [post]
func (c *User) GetRedEnvelope(ctx *gin.Context) {
}
