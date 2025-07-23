package handle

import (
	"context"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/async"
	"rk-api/pkg/cjson"
	"rk-api/pkg/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const QueueNotification = "notification"

func NewNotificationQueue(flow *entities.Notification) (*asynq.Task, error) {
	payload, err := cjson.Cjson.Marshal(flow)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(QueueNotification, payload, asynq.MaxRetry(1)), nil
}

func NotificationHandle(ctx context.Context, t *asynq.Task) error {
	var flow entities.Notification
	err := cjson.Cjson.Unmarshal(t.Payload(), &flow)
	if err != nil {
		return err
	}
	logger.Info("NotificationHandle:", string(t.Payload()))
	return nil
}

func NewNotificationHandler(srv async.IAsyncService) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var flow entities.Notification
		err := cjson.Cjson.Unmarshal(t.Payload(), &flow)
		if err != nil {
			return err
		}
		err = srv.HandleNotification(&flow)
		if err != nil {
			logger.ZError("Notification:",
				zap.Error(err),
			)
			return err
		}
		return nil
	}
}
