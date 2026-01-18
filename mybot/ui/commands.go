package ui

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStartCommand - обработка команды /start
func HandleStartCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	log.Printf("🚀 Команда /start от @%s", msg.From.UserName)
	
	// Используем универсальную функцию главного меню
	SendMainMenu(bot, msg.Chat.ID)
}

// HandleHelpCommand - обработка команды /help
func HandleHelpCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "📋 Команды:\n/start - Начать\n/help - Помощь\n/refresh_me - Обновить триггеры из БД"
	SendMessage(bot, msg.Chat.ID, text, "справка")
}
