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

const QueueInviteRelation = "invite:relation"

func NewInviteRelationQueue(relation *entities.HallInviteRelation) (*asynq.Task, error) {
	payload, err := cjson.Cjson.Marshal(relation)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(QueueInviteRelation, payload, asynq.MaxRetry(2)), nil
}

func InviteRelationHandle(ctx context.Context, t *asynq.Task) error {
	var relation entities.HallInviteRelation
	err := cjson.Cjson.Unmarshal(t.Payload(), &relation)
	if err != nil {
		return err
	}
	logger.Info("InviteRelationHandle:", string(t.Payload()))
	return nil
}

func NewInviteRelationHandle(srv async.IAsyncService) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var relation entities.HallInviteRelation
		err := cjson.Cjson.Unmarshal(t.Payload(), &relation)
		if err != nil {
			return err
		}

		// 在这里使用 FlowService 实例，flowSrv
		err = srv.FillInviteRelation(&relation)
		if err != nil {
			logger.ZError("InviteRelationHandle:",
				zap.String("relation", string(t.Payload())),
				zap.Error(err),
			)
		}

		return err
	}
}
