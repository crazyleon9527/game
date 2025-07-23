package task

import (
	"rk-api/internal/app/service/async"
	"rk-api/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

type ProcessStatsJob struct {
	Srv async.IAsyncService
}

var gProcessStatsLock sync.Mutex

func (r ProcessStatsJob) Run() {
	gProcessStatsLock.Lock()
	defer gProcessStatsLock.Unlock()
	err := r.Srv.MakeProfitRank() //
	if err != nil {
		logger.ZError("ProcessStatsJob", zap.Error(err))
		return
	}

}

type ProcessSyncThirdPartyDataJob struct {
	Srv async.IAsyncService
}

var gProcessSyncThirdPartyDataLock sync.Mutex

func (r ProcessSyncThirdPartyDataJob) Run() {
	gProcessSyncThirdPartyDataLock.Lock()
	defer gProcessSyncThirdPartyDataLock.Unlock()
	r.Srv.SyncThirdPartyData()
}
