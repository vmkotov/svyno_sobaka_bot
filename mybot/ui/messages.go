package ui

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendMessage - общая функция отправки сообщений
func SendMessage(bot *tgbotapi.BotAPI, chatID int64, text, context string) {
	reply := tgbotapi.NewMessage(chatID, text)

	if _, err := bot.Send(reply); err != nil {
		log.Printf("❌ Ошибка отправки %s: %v", context, err)
	} else {
		log.Printf("✅ Отправлен %s", context)
	}
}
