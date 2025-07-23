package handle

import (
	"context"

	"github.com/hibiken/asynq"
)

const QueueOrderExpiration = "order:expiration"

func NewOrderExpirationQueue(tradeId string) (*asynq.Task, error) {
	return asynq.NewTask(QueueOrderExpiration, []byte(tradeId)), nil
}

// OrderExpirationHandle 设置订单过期
func OrderExpirationHandle(ctx context.Context, t *asynq.Task) error {
	// tradeId := string(t.Payload())

	return nil
}
