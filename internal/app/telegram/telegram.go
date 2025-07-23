package telegram

import (
	"rk-api/internal/app/config"
	"time"

	"go.uber.org/zap"
	tb "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

//

var bot *tb.Bot

// bottart 机器人启动
func InitBot() (*tb.Bot, error) {
	var err error
	botetting := tb.Settings{
		Token:  config.Get().TelegramSetting.ApiToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}
	if config.Get().TelegramSetting.Proxy != "" {
		botetting.URL = config.Get().TelegramSetting.Proxy
	}
	bot, err = tb.NewBot(botetting)
	if err != nil {
		zap.L().Sugar().Error(err.Error())
		return nil, err
	}
	err = bot.SetCommands(Cmds)
	if err != nil {
		zap.L().Sugar().Error(err.Error())
		return nil, err
	}
	RegisterHandle()
	bot.Start()

	return bot, nil
}

// RegisterHandle 注册处理器
func RegisterHandle() {
	adminOnly := bot.Group()
	adminOnly.Use(middleware.Whitelist(config.Get().TelegramSetting.ManagerID))
	adminOnly.Handle(START_CMD, WalletList)
	adminOnly.Handle(tb.OnText, OnTextMessageHandle)
}

// SendToBot 主动发送消息机器人消息
func SendToManage(msg string) {
	go func() {
		user := tb.User{
			ID: config.Get().TelegramSetting.ManagerID,
		}
		_, err := bot.Send(&user, msg, &tb.SendOptions{
			ParseMode: tb.ModeHTML,
		})

		if err != nil {
			zap.L().Info("telebot", zap.Int64("chatID", config.Get().TelegramSetting.ManagerID), zap.String("msg", msg))
		}
	}()
}

func SendTo(chatID int64, msg string) {
	go func() {
		user := tb.User{
			ID: config.Get().TelegramSetting.ManagerID,
		}
		_, err := bot.Send(&user, msg, &tb.SendOptions{
			ParseMode: tb.ModeHTML,
		})
		if err != nil {
			zap.L().Info("telebot", zap.Int64("chatID", chatID), zap.String("msg", msg))
		}
	}()
}

// msgTpl := `
// <b>📢📢有新的交易支付成功！</b>
// <pre>交易号：%s</pre>
// <pre>订单号：%s</pre>
// <pre>请求支付金额：%f cny</pre>
// <pre>实际支付金额：%f usdt</pre>
// <pre>钱包地址：%s</pre>
// <pre>订单创建时间：%s</pre>
// <pre>支付成功时间：%s</pre>
// `
// msg := fmt.Sprintf(msgTpl, order.TradeId, order.OrderId, order.Amount, order.ActualAmount, order.Token, order.CreatedAt.ToDateTimeString(), carbon.Now().ToDateTimeString())
// telegram.SendToManage(msg)
