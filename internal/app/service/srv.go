package service

import (
	"rk-api/internal/app/entities"

	"github.com/google/wire"
)

// SrvSet repo injection
var SrvSet = wire.NewSet(
	StateServiceSet,
	AgentServiceSet,
	AdminServiceSet,
	FlowServiceSet,
	// LinkevoServiceSet,
	RechargeServiceSet,
	UserServiceSet,
	WingoServiceSet,
	NineServiceSet,
	WithdrawServiceSet,
	ActivityServiceSet,
	AsyncServiceManagerSet,
	QuizServiceSet,
	R8ServiceSet,
	ZfServiceSet,
	StatsServiceSet,
	GameServiceSet,
	OauthServiceSet,
	AuthServiceSet,
	VerifyServiceSet,
	JhszServiceSet,
	ChainServiceSet,
	WalletServiceSet,
	FinancialServiceSet,
	NotificationServiceSet,
	ChatServiceSet,
	PlatServiceSet,
	RealServiceSet,
	TransactionServiceSet,
	FreeServiceSet,
	HashGameServiceSet,
	CrashGameServiceSet,
	MineGameServiceSet,
	DiceGameServiceSet,
	LimboGameServiceSet,
) // end

var AsyncServiceManagerSet = wire.NewSet(wire.Struct(new(AsyncServiceManager), "*"))

type AsyncServiceManager struct {
	FlowSrv         *FlowService
	UserSrv         *UserService
	AgentSrv        *AgentService
	ActivitySrv     *ActivityService
	AdminSrv        *AdminService
	StatsSrv        *StatsService
	RechargeSrv     *RechargeService
	R8Srv           *R8Service
	ZfSrv           *ZfService
	WingoSrv        *WingoService
	NineSrv         *NineService
	ChainSrv        *ChainService
	NotificationSrv *NotificationService
	GameSrv         *GameService
}

//添加了 记得重新wire

func (m *AsyncServiceManager) CreateFlow(flow *entities.Flow) error {
	return m.FlowSrv.CreateFlow(flow)
}

func (m *AsyncServiceManager) ExpireUser(uid uint) error {
	return m.UserSrv.ClearUserCache(uid)
}

func (m *AsyncServiceManager) FillInviteRelation(relation *entities.HallInviteRelation) error {
	return m.AgentSrv.FillInviteRelation(relation)
}

func (m *AsyncServiceManager) SettleDailyInterest() {
	m.ActivitySrv.SettleDailyInterest()
}

func (m *AsyncServiceManager) InvitePinduo(relation *entities.HallInvitePinduo) error {
	return m.ActivitySrv.InvitePinduoOnce(relation)
}

func (m *AsyncServiceManager) CreateSystemOptionLog(entity *entities.SystemOptionLog) error { //添加系统操作日志
	return m.AdminSrv.CreateSystemOptionLog(entity)
}

func (m *AsyncServiceManager) ProcessRefundGameFlow(limit int) error {
	return m.FlowSrv.ProcessRefundGameFlow(limit)
}

func (m *AsyncServiceManager) MakeProfitRank() error { //生成盈利排行榜
	return m.StatsSrv.MakeProfitRank()
}

func (m *AsyncServiceManager) BatchFinalizeGameCashReturn(limit int) error { //领取游戏返现
	return m.AgentSrv.BatchFinalizeGameCashReturn(limit)
}

func (m *AsyncServiceManager) QueryAndUpdateRechargeChannelBalance() error { //查询支付平台的余额信息
	return m.RechargeSrv.QueryAndUpdateRechargeChannelBalance()
}
func (m *AsyncServiceManager) QueryR8BetRecords() error { //查询r8投注记录
	return m.R8Srv.QueryBetRecords()
}
func (m *AsyncServiceManager) QueryZFBetRecords() error { //查询r8投注记录
	return m.ZfSrv.QueryBetRecords()
}
func (m *AsyncServiceManager) SyncThirdPartyData() { //同步第三方游戏数据
	m.StatsSrv.SyncThirdPartyData()
}

func (m *AsyncServiceManager) SyncThirdOnlineCount() { //查询第三方在线人数
	m.GameSrv.SyncThirdOnlineCount()
}

func (m *AsyncServiceManager) QuerySettleExpiredWingos() error { //查询结算wingo
	return m.WingoSrv.QuerySettleExpiredWingos()
}
func (m *AsyncServiceManager) QuerySettleExpiredNines() error { //查询结算nine
	return m.NineSrv.QuerySettleExpiredNines()
}

func (m *AsyncServiceManager) MonthBackupAndClean(tableNames string) error { ////每月备份并清理数据
	return m.AdminSrv.CallMonthBackupAndClean(&entities.MonthBackupAndCleaReq{TableNames: tableNames, OptionID: 1})
}

func (m *AsyncServiceManager) HandleNotification(notification *entities.Notification) error { //处理通知
	return m.NotificationSrv.HandleNotification(notification)
}

// func (m *AsyncServiceManager) ProcessChainRetryTransaction(transation *entities.ChainTransaction, failed bool) error { //处理plat chain deposit
// 	return m.ChainSrv.ProcessChainRetryTransaction(transation, failed)
// }

//领取游戏返现

//pinduo邀请
