package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// HashGameSet 注入HashGame
var HashGameSet = wire.NewSet(wire.Struct(new(HashGame), "*"))

type HashGame struct {
}

// FairCheck 公平性检查
// @Summary 公平性检查
// @Description 公平性检查
// @Tags hash游戏
// @Produce json
// @Param req body entities.FairCheckReq true "params"
// @Success 200 {object} entities.FairCheckRsp "成功返回游戏结果"
// @Router /api/hashgame/fire-check [post]
func (c *HashGame) FairCheck(ctx *gin.Context) {
}
