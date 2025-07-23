package task

import (
	"rk-api/internal/app/service/async"
	"rk-api/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

type ProcessSettleExpiredWingoJob struct {
	Srv async.IAsyncService
}

var gProcessSettleExpiredWingoLock sync.Mutex

func (r ProcessSettleExpiredWingoJob) Run() {
	gProcessSettleExpiredWingoLock.Lock()
	defer gProcessSettleExpiredWingoLock.Unlock()
	err := r.Srv.QuerySettleExpiredWingos()
	if err != nil {
		logger.ZError("ProcessSettleExpiredWingoJob", zap.Error(err))
		return
	}
}

type ProcessSettleExpiredNineJob struct {
	Srv async.IAsyncService
}

var gProcessSettleExpiredNineLock sync.Mutex

func (r ProcessSettleExpiredNineJob) Run() {
	gProcessSettleExpiredNineLock.Lock()
	defer gProcessSettleExpiredNineLock.Unlock()
	err := r.Srv.QuerySettleExpiredNines()
	if err != nil {
		logger.ZError("ProcessSettleExpiredNineJob", zap.Error(err))
		return
	}

}
