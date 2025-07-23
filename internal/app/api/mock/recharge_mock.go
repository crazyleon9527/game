package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// RechargeSet 注入Recharge
var RechargeSet = wire.NewSet(wire.Struct(new(Recharge), "*"))

// Recharge 示例程序
type Recharge struct {
}

// @Tags Recharge
// @Summary 获取充值表
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.GetRechargeOrderListReq true "params"
// @Success 200 {object} entities.Paginator{List=[]entities.HallWithdrawRecord}
// @Router /api/recharge/get-recharge-order-list [post]
func (a *Flow) GetRechargeOrderList(c *gin.Context) {
}

// @Tags Recharge
// @Summary 获取充值渠道配置
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entities.RechargeConfig
// @Router /api/recharge/get-recharge-config [post]
func (a *Flow) GetRechargeConfig(c *gin.Context) {
}

// @Tags Recharge
// @Summary 获取充值地址信息
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.GetRechargeUrlReq true "params"
// @Success 200 {object} entities.RechargeUrlInfo
// @Router /api/recharge/get-recharge-url [post]
func (a *Flow) GetRechargeUrlInfo(c *gin.Context) {
}
