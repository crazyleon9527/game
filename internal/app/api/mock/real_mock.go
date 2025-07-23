package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// RealSet 注入Real
var RealSet = wire.NewSet(wire.Struct(new(Real), "*"))

type Real struct {
}

func (a *Real) GetRealSetting(c *gin.Context) {
}

// @Tags Real
// @Summary 提交实名认证信息
// @Description 提交实名认证信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entities.RealNameAuthReq
// @Router /api/real/commit-real-auth [post]
func (c *Real) CommitRealAuth(ctx *gin.Context) {

}

// @Tags Real
// @Summary 获取实名认证信息
// @Description 获取实名认证信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entities.RealAuth
// @Router /api/real/get-real-auth [post]
func (c *Real) GetRealAuth(ctx *gin.Context) {

}

// @Tags Real
// @Summary 更新实名认证信息
// @Description 更新实名认证信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entities.UpdateRealNameAuthReq
// @Router /api/real/update-real-auth [post]
func (c *Real) UpdateRealNameAuth(ctx *gin.Context) {

}
