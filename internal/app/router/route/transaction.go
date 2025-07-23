package route

import (
	"github.com/gin-gonic/gin"
	"rk-api/internal/app/api"
	"rk-api/internal/app/middleware"
)

func RegisterTransactionRoutes(r *gin.RouterGroup, txnAPI *api.TransactionAPI) {
	txn := r.Group("/transaction")
	{
		txn.POST("/get-transaction-list", middleware.JWTMiddleware(), txnAPI.GetTransactionList)
	}
}
