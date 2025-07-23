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

// func NewImageResizeTask(src string) (*asynq.Task, error) {
//     payload, err := json.Marshal(ImageResizePayload{SourceURL: src})
//     if err != nil {
//         return nil, err
//     }
//     // task options can be passed to NewTask, which can be overridden at enqueue time.
//     return asynq.NewTask(TypeImageResize, payload, asynq.MaxRetry(5), asynq.Timeout(20 * time.Minute)), nil
// }

const QueueInvitePinduo = "invite:pinduo"

func NewInvitePinduoQueue(relation *entities.HallInvitePinduo) (*asynq.Task, error) {
	payload, err := cjson.Cjson.Marshal(relation)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(QueueInvitePinduo, payload, asynq.MaxRetry(1)), nil
}

func InvitePinduoHandle(ctx context.Context, t *asynq.Task) error {
	var relation entities.HallInvitePinduo
	err := cjson.Cjson.Unmarshal(t.Payload(), &relation)
	if err != nil {
		return err
	}
	logger.Info("InvitePinduoHandle:", string(t.Payload()))
	return nil
}

func NewInvitePinduoHandle(srv async.IAsyncService) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var relation entities.HallInvitePinduo
		err := cjson.Cjson.Unmarshal(t.Payload(), &relation)
		if err != nil {
			return err
		}
		err = srv.InvitePinduo(&relation)
		if err != nil {
			logger.ZError("InvitePinduoHandle:",
				zap.String("relation", string(t.Payload())),
				zap.Error(err),
			)
		}
		return nil
	}
}
