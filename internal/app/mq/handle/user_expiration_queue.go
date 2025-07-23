package handle

import (
	"context"
	"rk-api/internal/app/service/async"
	"rk-api/pkg/logger"
	"strconv"

	"github.com/hibiken/asynq"
)

const QueueUserExpiration = "user:expiration"

func NewUserExpirationQueue(uid string) (*asynq.Task, error) {
	return asynq.NewTask(QueueUserExpiration, []byte(uid), asynq.MaxRetry(3)), nil
}

func UserExpirationHandle(ctx context.Context, t *asynq.Task) error {
	uid := string(t.Payload())
	userID, _ := strconv.ParseUint(uid, 10, 64)
	logger.Info("UserExpirationHandle", userID)
	return nil
}

func NewUserExpirationHandle(userSrv async.IAsyncService) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		uid := string(t.Payload())
		userID, _ := strconv.ParseUint(uid, 10, 64)
		logger.Info("UserExpirationHandle", userID)
		return userSrv.ExpireUser(uint(userID))
	}
}
