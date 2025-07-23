package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// 使用环境变量获取 Telegram API Token
	bot, err := tgbotapi.NewBotAPI("7723141080:AAHLiMtnaGkcoT0xApmLC9dxN3Wqzpi-l6I")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)
	log.Printf("Start listening...")

	for update := range updates {
		log.Printf(" %s", update.CallbackQuery.GameShortName)
		if update.CallbackQuery != nil && update.CallbackQuery.GameShortName == "sayMeLandLord" {
			// 创建回调响应配置
			callbackConfig := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			callbackConfig.URL = "http://h5-platform.jhkj.ddns.us/#/" // 设置游戏 URL

			// 发送响应
			if _, err := bot.Request(callbackConfig); err != nil {
				log.Printf("Error answering callback: %v", err)
			}
		}
	}

	// for update := range updates {
	// 	log.Println("==========================", update.Message)
	// 	if update.Message == nil { // 忽略非消息更新
	// 		continue
	// 	}

	// 	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	// 	// 处理命令
	// 	if update.Message.IsCommand() {
	// 		switch update.Message.Command() {
	// 		case "start":
	// 			replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "欢迎使用游戏机器人！输入 /play 开始游戏。")
	// 			if _, err := bot.Send(replyMsg); err != nil {
	// 				log.Panic(err)
	// 			}
	// 			fmt.Println("start command received")
	// 		case "play":
	// 			gameURL := "http://h5-platform.jhkj.ddns.us/#/" // 替换为您的游戏 URL

	// 			fmt.Println("play command received")

	// 			// 创建内联键盘
	// 			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
	// 				tgbotapi.NewInlineKeyboardRow(
	// 					tgbotapi.NewInlineKeyboardButtonURL("启动游戏", gameURL),
	// 				),
	// 			)

	// 			replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "点击下面的链接开始游戏！")
	// 			replyMsg.ReplyMarkup = inlineKeyboard

	// 			if _, err := bot.Send(replyMsg); err != nil {
	// 				log.Panic(err)
	// 			}
	// 		default:
	// 			replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "未知命令，请使用 /start 或 /play。")
	// 			if _, err := bot.Send(replyMsg); err != nil {
	// 				log.Panic(err)
	// 			}
	// 		}
	// 	} else {
	// 		replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "请使用 /start 或 /play 命令。")
	// 		if _, err := bot.Send(replyMsg); err != nil {
	// 			log.Panic(err)
	// 		}
	// 	}
	// }
}
