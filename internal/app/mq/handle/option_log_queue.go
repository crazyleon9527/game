package handle

import (
	"context"
	"rk-api/internal/app/entities"
	"rk-api/internal/app/service/async"
	"rk-api/pkg/cjson"
	"rk-api/pkg/logger"
	"strconv"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const QueueOptionLog = "option:log"

func NewOptionLogQueue(log *entities.SystemOptionLog) (*asynq.Task, error) {
	payload, err := cjson.Cjson.Marshal(log)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(QueueOptionLog, payload, asynq.MaxRetry(1)), nil
}

func OptionLogHandle(ctx context.Context, t *asynq.Task) error {
	uid := string(t.Payload())
	userID, _ := strconv.ParseUint(uid, 10, 64)
	logger.Info("OptionLogHandle", userID)
	return nil
}

func NewOptionLogHandle(srv async.IAsyncService) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var log entities.SystemOptionLog
		err := cjson.Cjson.Unmarshal(t.Payload(), &log)
		if err != nil {
			logger.ZError("CreateSystemOptionLog:",
				zap.String("log", string(t.Payload())),
				zap.Error(err),
			)
			return err
		}

		return srv.CreateSystemOptionLog(&log)
	}
}
