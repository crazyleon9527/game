package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// VerifySet 注入Verify
var VerifySet = wire.NewSet(wire.Struct(new(Verify), "*"))

// Verify 示例程序
type Verify struct {
}

// @Tags Verify
// @Summary 发送验证码
// @Accept  json
// @Produce  json
// @Param req body entities.VerifyCodeReq true "params"
// @Success 200 {object} ginx.Resp
// @Router /api/verify/send-verify-code [post]
func (c *Verify) SendVerifyCode(ctx *gin.Context) {
}
