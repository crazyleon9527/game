package task

import (
	"rk-api/internal/app/service/async"
	"rk-api/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

type ProcessQueryPlatBalanceJob struct {
	Srv async.IAsyncService
}

var gProcessQueryPlatBalanceLock sync.Mutex

func (r ProcessQueryPlatBalanceJob) Run() {
	gProcessQueryPlatBalanceLock.Lock()
	defer gProcessQueryPlatBalanceLock.Unlock()
	err := r.Srv.QueryAndUpdateRechargeChannelBalance()

	if err != nil {
		logger.ZError("ProcessQueryPlatBalanceJob", zap.Error(err))
		return
	}
}
