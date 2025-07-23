package handle

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"rk-api/internal/app/entities"
// 	"rk-api/internal/app/service/async"
// 	"rk-api/pkg/cjson"
// 	"rk-api/pkg/logger"

// 	"github.com/hibiken/asynq"
// 	"go.uber.org/zap"
// )

// const QueueChainDeposit = "chain:deposit"

// func NewChainDepositQueue(transaction *entities.ChainTransaction) (*asynq.Task, error) {
// 	payload, err := cjson.Cjson.Marshal(transaction)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return asynq.NewTask(QueueChainDeposit, payload, asynq.MaxRetry(5)), nil
// }

// func NewChainDepositHandler(srv async.IAsyncService) asynq.HandlerFunc {
// 	return func(ctx context.Context, t *asynq.Task) error {
// 		var transaction entities.ChainTransaction
// 		if err := json.Unmarshal(t.Payload(), &transaction); err != nil {
// 			// 正确做法：解析失败时跳过重试
// 			return fmt.Errorf("json.Unmarshal failed: %w", asynq.SkipRetry)
// 		}

// 		// 正确获取重试次数（仅用于日志记录）
// 		retryCount, _ := asynq.GetRetryCount(ctx)
// 		maxRetry, _ := asynq.GetMaxRetry(ctx) // 获取最大重试次数

// 		err := srv.ProcessChainRetryTransaction(&transaction, retryCount >= maxRetry)
// 		if err != nil {
// 			logger.ZError("ChainDepositHandle:",
// 				zap.String("transaction", string(t.Payload())),
// 				zap.Int("retry_count", retryCount), // 记录当前重试次数
// 				zap.Error(err),
// 			)
// 			return err // 触发自动重试
// 		}
// 		return nil
// 	}
// }
