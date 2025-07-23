package mock

import "github.com/google/wire"

// MockSet 注入mock
var MockSet = wire.NewSet(
	DemoSet,
	AgentSet,
	FlowSet,
	RechargeSet,
	UserSet,
	WingoSet,
	WithdrawSet,
	GameSet,
	OauthSet,
	AuthSet,
	VerifySet,
	ChainSet,
	WalletSet,
	NotificationSet,
	TransactionSet,
	StatsSet,
	HashGameSet,
	SDGameSet,
	CrashGameSet,
	MineGameSet,
	DiceGameSet,
	LimboGameSet,
)
