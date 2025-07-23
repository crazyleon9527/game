package async

import "rk-api/internal/app/entities"

type IAsyncService interface {
	CreateFlow(flow *entities.Flow) error                           //创建流水
	ExpireUser(uid uint) error                                      //让用户redis过期
	FillInviteRelation(relation *entities.HallInviteRelation) error //填充好友关系
	ProcessRefundGameFlow(limit int) error                          //处理游戏返利
	SettleDailyInterest()
	InvitePinduo(relation *entities.HallInvitePinduo) error       //pinduo邀请
	CreateSystemOptionLog(entity *entities.SystemOptionLog) error //添加系统操作日志
	MakeProfitRank() error                                        //生成盈利排行榜
	BatchFinalizeGameCashReturn(limit int) error                  //领取游戏返现
	QueryAndUpdateRechargeChannelBalance() error                  //查询更新支付平台余额信息

	QueryR8BetRecords() error //查询R8投注记录
	QueryZFBetRecords() error //查询ZF投注记录
	SyncThirdPartyData()      //同步第三方游戏数据
	SyncThirdOnlineCount()    // 同步第三方在线人数

	QuerySettleExpiredWingos() error             //查询结算wingo
	QuerySettleExpiredNines() error              //查询结算nine
	MonthBackupAndClean(tableNames string) error //每月备份并清理数据

	HandleNotification(notification *entities.Notification) error //处理通知

	// ProcessChainRetryTransaction(transation *entities.ChainTransaction, failed bool) error //上交易
}
