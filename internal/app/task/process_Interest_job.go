package task

import (
	"rk-api/internal/app/service/async"
	"sync"
)

type ProcessInterestJob struct {
	Srv async.IAsyncService
}

var gProcessInterestLock sync.Mutex

func (r ProcessInterestJob) Run() {
	gProcessInterestLock.Lock()
	defer gProcessInterestLock.Unlock()
	r.Srv.SettleDailyInterest()

}
