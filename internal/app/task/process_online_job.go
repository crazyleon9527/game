package task

import (
	"rk-api/internal/app/service/async"
	"rk-api/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

type ProcessOnlineJob struct {
	Srv async.IAsyncService
}

var gProcessOnlineLock sync.Mutex

func (r ProcessOnlineJob) Run() {
	gProcessOnlineLock.Lock()
	defer gProcessOnlineLock.Unlock()
	err := r.Srv.MakeProfitRank() //
	if err != nil {
		logger.ZError("ProcessOnlineJob", zap.Error(err))
		return
	}

}

type ProcessSyncThirdOnlineCountJob struct {
	Srv async.IAsyncService
}

var gProcessSyncThirdOnlineCountLock sync.Mutex

func (r ProcessSyncThirdOnlineCountJob) Run() {
	gProcessSyncThirdOnlineCountLock.Lock()
	defer gProcessSyncThirdOnlineCountLock.Unlock()
	r.Srv.SyncThirdOnlineCount()
}
