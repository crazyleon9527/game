package task

import (
	"rk-api/internal/app/config"
	"rk-api/internal/app/service/async"
	"time"

	"github.com/robfig/cron/v3"
)

func Start(service async.IAsyncService) error { //当格式不正确的时候会返回error
	// c := cron.New()

	// 创建一个新的cron实例，并设定时区为东八区，即北京时间
	// 设置时区为东八区，即北京时间
	loc, err := time.LoadLocation(config.Get().ServiceSettings.Timezone)
	if err != nil {
		return err
	}
	// 创建新的cron实例，设置时区，并添加错误恢复机制
	c := cron.New(
		cron.WithLocation(loc),
		cron.WithChain(
			cron.Recover(cron.DefaultLogger), // 使用内建的恢复机制，避免panic导致程序崩溃
		),
	)

	_, err = c.AddJob("@every 2m", ProcessSyncThirdPartyDataJob{Srv: service}) //同步第三方游戏数据
	if err != nil {
		return err // 返回错误而不是结束程序
	}

	// _, err = c.AddJob("@every 1m", ProcessSyncThirdOnlineCountJob{Srv: service}) //同步第三方游戏在线人数
	// if err != nil {
	// 	return err // 返回错误而不是结束程序
	// }

	// _, err = c.AddJob("1 0 * * *", ProcessInterestJob{Srv: service}) //利息   零点 1分
	// _, err = c.AddJob("1 0 * * *", ProcessInterestJob{Srv: service})
	// if err != nil {
	// 	return err // 返回错误而不是结束程序
	// }

	// _, err = c.AddJob("@every 20s", ProcessRefundJob{Srv: service}) //游戏返利  现在是直接拉取返利
	// if err != nil {
	// 	return err //
	// }

	// _, err = c.AddJob("@every 5m", ProcessQueryPlatBalanceJob{Srv: service}) //查询支付平台金额
	// if err != nil {
	// 	return err //
	// }

	// _, err = c.AddJob("@every 3m", ProcessSettleExpiredWingoJob{Srv: service}) //处理超时未结算的wingo订单
	// if err != nil {
	// 	return err //
	// }

	// _, err = c.AddJob("@every 3m", ProcessSettleExpiredNineJob{Srv: service}) //处理超时未结算的nine订单
	// if err != nil {
	// 	return err //
	// }

	_, err = c.AddJob("1 0 1 * *", ProcessBackupCleanGameReturnJob{Srv: service}) //每个月的游戏返利备份清理  每个月1号的 凌晨 过 5分钟
	if err != nil {
		return err // 返回错误而不是结束程序
	}

	_, err = c.AddJob("5 0 1 * *", ProcessBackupCleanFlowJob{Srv: service}) //每个月的流水备份清理
	if err != nil {
		return err // 返回错误而不是结束程序
	}

	_, err = c.AddJob("30 1 1 * *", ProcessBackupCleanCrashGameRoundJob{Srv: service}) //每个月的崩溃游戏轮次备份清理
	if err != nil {
		return err // 返回错误而不是结束程序
	}

	// _, err = c.AddJob("10 0 1 * *", ProcessBackupCleanRefundFlowJob{Srv: service}) //每个月的返利流水备份清理
	// if err != nil {
	// 	return err // 返回错误而不是结束程序
	// }

	// _, err = c.AddJob("20 0 2 * *", ProcessBackupCleanRefundLinkGameFlowJob{Srv: service}) //每个月的返利link游戏流水备份清理
	// if err != nil {
	// 	return err // 返回错误而不是结束程序
	// }

	// _, err = c.AddJob("30 0 1 * *", ProcessBackupCleanGameOrderJob{Srv: service}) //每个月的游戏订单备份清理
	// if err != nil {
	// 	return err // 返回错误而不是结束程序
	// }

	// _, err = c.AddJob("40 2 1 * *", ProcessBackupCleanLinkGameOrderJob{Srv: service}) //每个月的外链游戏订单备份清理
	// if err != nil {
	// 	return err // 返回错误而不是结束程序
	// }

	// logger.ZPanic("test-------panic")

	// _, err = c.AddJob("@every 59m", ProcessR8BetRecordJob{Srv: service}) //每个小时处理 r8投注记录
	// if err != nil {
	// 	return err //
	// }

	// _, err = c.AddJob("@every 1h", ProcessZfBetRecordJob{Srv: service}) //每个小时处理 zf投注记录
	// if err != nil {
	// 	return err //
	// }

	// _, err = c.AddJob("@every 10m", ProcessGetGameReturnCashJob{Srv: service}) //自动领取游戏返利
	//

	// if err != nil {
	// 	return err // 返回错误而不是结束程序
	// }

	go c.Start() // 在一个新的goroutine中启动cron调度器
	return nil
}

// 时间表达式"5 0 1 * *" 表示每个月的1号的0点5分执行。
// 具体解释如下：
// 第一个位（分钟）：表示每小时的哪一分钟触发任务，这里是5。
// 第二个位（小时）：表示每天的哪个小时触发任务，这里是0。
// 第三个位（日期）：表示每月的哪一天触发任务，这里是1。
// 第四个位（月份）：表示每年的哪个月触发任务，这里是任意月，使用通配符*表示。
// 第五个位（星期）：表示每周的哪一天触发任务，这里也是任意星期，使用通配符*表示。
// 因此，时间表达式"5 0 1 * *" 表示在每个月的1号的0点5分触发任务。

// // 自定义的恢复中间件
// func CustomRecover(handler cron.Job) cron.Job {
//     return cron.FuncJob(func() {
//         defer func() {
//             // 使用recover()检查是否有panic发生
//             if r := recover(); r != nil {
//                 // 可以在这里添加您自己的逻辑，比如日志记录
//                 log.Printf("Recovered from a panic in job. Error: %v", r)
//                 // 这里也可以添加发送警报的代码，比如发送邮件通知等
//             }
//         }()
//         handler.Run() // 执行原本的任务
//     })
// }

// // 创建新的cron实例，并使用自定义的恢复中间件
// c := cron.New(cron.WithChain(CustomRecover))
