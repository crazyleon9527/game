package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// FlowSet 注入Flow
var FlowSet = wire.NewSet(wire.Struct(new(Flow), "*"))

// Flow 示例程序
type Flow struct {
}

// @Tags Flow
// @Summary 获取流水列表
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.GetFlowListReq true "params"
// @Success 200 {object} entities.Paginator{List=[]entities.Flow}
// @Router /api/flow/get-flow-list [post]
func (a *Flow) GetFlowList(c *gin.Context) {
}
