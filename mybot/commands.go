package mybot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCommand - определяет команду и вызывает соответствующий обработчик
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		handleStartCommand(bot, msg)
	case "help":
		handleHelpCommand(bot, msg)
		// Можно добавить другие команды
	}
}
