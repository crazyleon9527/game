package mock

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// TransactionSet 注入Transaction
var TransactionSet = wire.NewSet(wire.Struct(new(Transaction), "*"))

type Transaction struct {
}

// @Tags Transaction
// @Summary 获取交易记录列表
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param req body entities.GetTransactionListReq true "params"
// @Success 200 {object}  entities.Transaction
// @Router /api/transaction/get-transaction-list [post]
func (a *Transaction) GetTransactionList(c *gin.Context) {
}
