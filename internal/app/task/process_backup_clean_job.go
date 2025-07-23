package task

import (
	"rk-api/internal/app/service/async"
	"rk-api/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

type ProcessBackupCleanGameReturnJob struct {
	Srv async.IAsyncService
}

var gProcessBackupCleanGameReturnLock sync.Mutex

func (r ProcessBackupCleanGameReturnJob) Run() {
	gProcessBackupCleanGameReturnLock.Lock()
	defer gProcessBackupCleanGameReturnLock.Unlock()
	err := r.Srv.MonthBackupAndClean("game_return") //() //
	if err != nil {
		logger.ZError("ProcessBackupCleanGameReturnJob", zap.Error(err))
		return
	}
}

type ProcessBackupCleanFlowJob struct {
	Srv async.IAsyncService
}

var gProcessBackupCleanFlowLock sync.Mutex

func (r ProcessBackupCleanFlowJob) Run() {
	gProcessBackupCleanFlowLock.Lock()
	defer gProcessBackupCleanFlowLock.Unlock()
	err := r.Srv.MonthBackupAndClean("flow") //() //
	if err != nil {
		logger.ZError("ProcessBackupCleanFlowJob", zap.Error(err))
		return
	}
}

type ProcessBackupCleanCrashGameRoundJob struct {
	Srv async.IAsyncService
}

var gProcessBackupCleanCrashGameRoundLock sync.Mutex

func (r ProcessBackupCleanCrashGameRoundJob) Run() {
	gProcessBackupCleanCrashGameRoundLock.Lock()
	defer gProcessBackupCleanCrashGameRoundLock.Unlock()
	err := r.Srv.MonthBackupAndClean("crash_game_round") //() //
	if err != nil {
		logger.ZError("ProcessBackupCleanCrashGameRoundJob", zap.Error(err))
		return
	}
}

type ProcessBackupCleanRefundFlowJob struct {
	Srv async.IAsyncService
}

var gProcessBackupCleanRefundFlowLock sync.Mutex

func (r ProcessBackupCleanRefundFlowJob) Run() {
	gProcessBackupCleanRefundFlowLock.Lock()
	defer gProcessBackupCleanRefundFlowLock.Unlock()
	err := r.Srv.MonthBackupAndClean("refund_game_flow") //() //
	if err != nil {
		logger.ZError("ProcessBackupCleanRefundFlowJob", zap.Error(err))
		return
	}
}

type ProcessBackupCleanRefundLinkGameFlowJob struct {
	Srv async.IAsyncService
}

var gProcessBackupCleanRefundLinkGameFlowLock sync.Mutex

func (r ProcessBackupCleanRefundLinkGameFlowJob) Run() {
	gProcessBackupCleanRefundLinkGameFlowLock.Lock()
	defer gProcessBackupCleanRefundLinkGameFlowLock.Unlock()
	err := r.Srv.MonthBackupAndClean("refund_link_game_flow") //() //
	if err != nil {
		logger.ZError("ProcessBackupCleanRefundLinkGameFlowJob", zap.Error(err))
		return
	}
}

// -- CALL MonthBackupAndCleanV4('wingo_order,nine_order');
type ProcessBackupCleanGameOrderJob struct {
	Srv async.IAsyncService
}

var gProcessBackupCleanGameOrderLock sync.Mutex

func (r ProcessBackupCleanGameOrderJob) Run() {
	gProcessBackupCleanGameOrderLock.Lock()
	defer gProcessBackupCleanGameOrderLock.Unlock()
	err := r.Srv.MonthBackupAndClean("wingo_order,nine_order") //() //
	if err != nil {
		logger.ZError("ProcessBackupCleanGameOrderJob", zap.Error(err))
		return
	}
}

// -- CALL MonthBackupAndCleanV4('r8_transfer_order');
// -- CALL MonthBackupAndCleanV4('zf_transfer_order');

type ProcessBackupCleanLinkGameOrderJob struct {
	Srv async.IAsyncService
}

var gProcessBackupCleanLinkGameOrderLock sync.Mutex

func (r ProcessBackupCleanLinkGameOrderJob) Run() {
	gProcessBackupCleanLinkGameOrderLock.Lock()
	defer gProcessBackupCleanLinkGameOrderLock.Unlock()
	err := r.Srv.MonthBackupAndClean("r8_transfer_order,zf_transfer_order") //() //
	if err != nil {
		logger.ZError("ProcessBackupCleanLinkGameOrderJob", zap.Error(err))
		return
	}
}
