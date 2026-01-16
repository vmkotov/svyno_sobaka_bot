package mybot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	forwarder "github.com/vmkotov/telegram-forwarder"
)

// forwardMessage - пересылает сообщение в указанный чат
func forwardMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, forwardChatID int64) {
	if forwardChatID == 0 {
		return // Пересылка отключена
	}

	forwarder.JustForward(bot, msg, forwardChatID)
	log.Printf("📤 Сообщение переслано в чат %d", forwardChatID)
}
