package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// WithdrawSet 注入Withdraw
var WithdrawSet = wire.NewSet(wire.Struct(new(Withdraw), "*"))

// Withdraw 示例程序
type Withdraw struct {
}

// @Tags Withdraw
// @Summary 获取提现详情
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entities.Paginator{List=[]entities.WithdrawDetail}
// @Router /api/withdraw/get-withdraw-detail [post]
func (c *Withdraw) GetWithdrawDetail(ctx *gin.Context) {

}

// @Tags Withdraw
// @Summary 获取玩家提现银行卡列表
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entities.Paginator{List=[]entities.WithdrawCard}
// @Router /api/withdraw/get-withdraw-card-list [post]
func (c *Withdraw) GetithdrawCardList(ctx *gin.Context) {

}

// @Tags Withdraw
// @Summary 获取玩家提现记录
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entities.Paginator{List=[]entities.HallWithdrawRecord}
// @Router /api/withdraw/get-withdraw-record-list [post]
func (c *Withdraw) GetHallWidthdrawRecordList(ctx *gin.Context) {

}

// @Tags Withdraw
// @Summary 申请提现
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.ApplyForWithdrawalReq true "params"
// @Success 200 {object} ginx.Resp
// @Router /api/withdraw/apply-for-withdrawal [post]
func (c *Withdraw) ApplyForWithdrawal(ctx *gin.Context) {

}

// @Tags Withdraw
// @Summary 添加提现银行卡
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.AddWithdrawCardReq true "params"
// @Success 200 {object} ginx.Resp
// @Router /api/withdraw/add-withdraw-card [post]
func (c *Withdraw) AddWithdrawCard(ctx *gin.Context) {

}

// @Tags Withdraw
// @Summary 删除银行卡
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.IDReq true "params"
// @Success 200 {object} ginx.Resp
// @Router /api/withdraw/del-withdraw-card [post]
func (c *Withdraw) DelWithdrawCard(ctx *gin.Context) {

}

// @Tags Withdraw
// @Summary 选择需要使用的银行卡
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.IDReq true "params"
// @Success 200 {object} ginx.Resp
// @Router /api/withdraw/select-withdraw-card [post]
func (c *Withdraw) SelectWithdrawCard(ctx *gin.Context) {

}
