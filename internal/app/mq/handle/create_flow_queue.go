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

const QueueCreateFlow = "flow:create"

func NewCreateFlowQueue(flow *entities.Flow) (*asynq.Task, error) {
	payload, err := cjson.Cjson.Marshal(flow)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(QueueCreateFlow, payload, asynq.MaxRetry(3)), nil
}

func CreateFlowHandle(ctx context.Context, t *asynq.Task) error {
	var flow entities.Flow
	err := cjson.Cjson.Unmarshal(t.Payload(), &flow)
	if err != nil {
		return err
	}
	logger.Info("CreateFlowHandle:", string(t.Payload()))
	return nil
}

func NewCreateFlowHandler(srv async.IAsyncService) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var flow entities.Flow
		err := cjson.Cjson.Unmarshal(t.Payload(), &flow)
		if err != nil {
			return err
		}
		err = srv.CreateFlow(&flow)
		if err != nil {
			logger.ZError("CreateFlow:",
				zap.String("flow", string(t.Payload())),
				zap.Error(err),
			)
			return err
		}
		return nil
	}
}
