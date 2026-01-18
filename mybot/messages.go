package mybot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStartCommand - обработка команды /start
func handleStartCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	log.Printf("🚀 Команда /start от @%s", msg.From.UserName)
	
	// Используем универсальную функцию главного меню
	sendMainMenu(bot, msg.Chat.ID)
}

// handleHelpCommand - обработка команды /help
func handleHelpCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "📋 Команды:\n/start - Начать\n/help - Помощь\n/refresh_me - Обновить триггеры из БД"
	sendMessage(bot, msg.Chat.ID, text, "справка")
}

// sendMessage - общая функция отправки сообщений
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text, context string) {
	reply := tgbotapi.NewMessage(chatID, text)

	if _, err := bot.Send(reply); err != nil {
		log.Printf("❌ Ошибка отправки %s: %v", context, err)
	} else {
		log.Printf("✅ Отправлен %s", context)
	}
}
