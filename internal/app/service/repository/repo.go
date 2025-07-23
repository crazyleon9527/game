package repository

import (
	"rk-api/internal/app/config"
	"rk-api/internal/app/entities"
	"rk-api/pkg/logger"
	"strings"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RepoSet repo injection
var RepoSet = wire.NewSet(
	AgentRepositorySet,
	AdminRepositorySet,
	RechargeRepositorySet,
	UserRepositorySet,
	FlowRepositorySet,
	WingoRepositorySet,
	NineRepositorySet,
	WithdrawRepositorySet,
	ActivityRepositorySet,
	QuizRepositorySet,
	R8RepositorySet,
	ZfRepositorySet,
	StatsRepositorySet,
	GameRepositorySet,
	OauthRepositorySet,
	JhszRepositorySet,
	ChainRepositorySet,
	WalletRepositorySet,
	FinancialRepositorySet,
	NotificationRepositorySet,
	ChatRepositorySet,
	PlatRepositorySet,
	RealRepositorySet,
	TransactionRepositorySet,
	FreeRepositorySet,
	HashGameRepositorySet,
	CrashGameRepositorySet,
	MineGameRepositorySet,
	DiceGameRepositorySet,
	LimboGameRepositorySet,
) // end

// Auto migration for given models
func AutoMigrate(db *gorm.DB) error {
	if dbType := config.Get().DBSettings.Driver; strings.ToLower(dbType) == "mysql" {
		db = db.Set("gorm:table_options", "ENGINE=InnoDB")
	}

	err := db.AutoMigrate(

		new(entities.HallInviteRelation),
		new(entities.GameReturn),
		new(entities.RechargeReturn),
		new(entities.RakeBack),

		new(entities.Flow),
		new(entities.RefundGameFlow),
		new(entities.RefundLinkGameFlow),

		new(entities.RechargeGood),
		new(entities.RechargeOrder),
		new(entities.RechargeChannelSetting),
		new(entities.RechargeSetting),
		new(entities.CompletedRecharge),

		new(entities.User),
		new(entities.VerifyCode),
		new(entities.HongbaoSetting),
		new(entities.HongbaoRecord),

		new(entities.WingoRoomSetting),
		new(entities.WingoPeriod),
		new(entities.WingoOrder),

		new(entities.NineRoomSetting),
		new(entities.NinePeriod),
		new(entities.NineOrder),

		new(entities.WithdrawCard),
		new(entities.HallWithdrawRecord),
		new(entities.CompletedWithdraw),
		new(entities.InterestCalculationLog),
		new(entities.PinduoSetting),
		new(entities.PinduoRecord),

		new(entities.R8TransferOrder),
		new(entities.R8ActivityOrder),
		new(entities.ZfTransferOrder),
		new(entities.R8BetRecord),
		new(entities.ZfBetRecord),

		new(entities.RankStats),
		new(entities.Game),

		new(entities.QuizFetchRule),
		new(entities.QuizEvent),
		new(entities.QuizMarket),
		new(entities.QuizBuyRecord),

		new(entities.FundFreeze),
		new(entities.FinancialSummary),
		new(entities.UserWallet),
		new(entities.Notification),
		new(entities.NotificationTemplate),
		new(entities.WalletAddress),
		new(entities.BlockchainToken),
		new(entities.PlatSetting),
		new(entities.ChatChannel),
		new(entities.ChatMessage),
		new(entities.GameRebateReceipt),
		new(entities.Activity),
		new(entities.Banner),
		new(entities.Logo),
		new(entities.GamerDailyStats),
		new(entities.GlobalSyncStatus),
		new(entities.GameRecord),

		new(entities.HashSDGameRound),
		new(entities.HashSDGameOrder),
		new(entities.CrashGameRound),
		new(entities.CrashGameOrder),
		new(entities.CrashAutoBet),
		new(entities.MineGameOrder),
		new(entities.DiceGameOrder),
	) // end

	AutoIncrement(db)
	return err
}

func AutoIncrement(db *gorm.DB) {
	sql := "ALTER TABLE `user` AUTO_INCREMENT = 8000000;" // 设置自增起始值的 SQL 语句
	result := db.Exec(sql)
	if result.Error != nil {
		// 处理错误
		logger.ZError("AutoIncrement user", zap.Error(result.Error))
	}
}

// ALTER TABLE `table_name` DROP INDEX `index_name`;

// gm_list
// hongbao_setting
// nine_room_setting
//pinduo_setting
//rake_back
//recharge_setting
//recharge_channel_setting
// /recharge_good
//sys_config
//sys_hooks
//sys_module2
//sys_option_log
//sys_url_route
//sys_user
//sys_user_admin
//sys_user_group
//sys_website
//wingo_room_setting
