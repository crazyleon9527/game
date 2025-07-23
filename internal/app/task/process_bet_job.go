package task

import (
	"rk-api/internal/app/service/async"
	"rk-api/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

type ProcessR8BetRecordJob struct {
	Srv async.IAsyncService
}

var gProcessR8BetRecordLock sync.Mutex

func (r ProcessR8BetRecordJob) Run() {
	gProcessR8BetRecordLock.Lock()
	defer gProcessR8BetRecordLock.Unlock()
	err := r.Srv.QueryR8BetRecords()
	if err != nil {
		logger.ZError("ProcessR8BetRecordJob", zap.Error(err))
		return
	}
}

type ProcessZfBetRecordJob struct {
	Srv async.IAsyncService
}

var gProcessZfBetRecordLock sync.Mutex

func (r ProcessZfBetRecordJob) Run() {
	gProcessZfBetRecordLock.Lock()
	defer gProcessZfBetRecordLock.Unlock()
	err := r.Srv.QueryZFBetRecords()
	if err != nil {
		logger.ZError("ProcessZfBetRecordJob", zap.Error(err))
		return
	}

}
