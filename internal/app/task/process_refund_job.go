package task

import (
	"rk-api/internal/app/service/async"
	"rk-api/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

type ProcessRefundJob struct {
	Srv async.IAsyncService
}

var gProcessRefundLock sync.Mutex

func (r ProcessRefundJob) Run() {

	gProcessRefundLock.Lock()
	defer gProcessRefundLock.Unlock()
	err := r.Srv.ProcessRefundGameFlow(90) //一次处理90条流水
	if err != nil {
		logger.ZError("ProcessRefundJob fail", zap.Error(err))
		return
	}
}

///////////////////////////////////////////////////////

type ProcessGetGameReturnCashJob struct {
	Srv async.IAsyncService
}

var gProcessGetGameReturnCashLock sync.Mutex

func (r ProcessGetGameReturnCashJob) Run() {
	gProcessGetGameReturnCashLock.Lock()
	defer gProcessGetGameReturnCashLock.Unlock()
	err := r.Srv.BatchFinalizeGameCashReturn(10) //
	if err != nil {
		logger.ZError("ProcessRefundJob", zap.Error(err))
		return
	}
}
