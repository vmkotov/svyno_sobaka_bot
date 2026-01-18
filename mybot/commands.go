package mybot

import (
	"database/sql"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCommand - определяет команду и вызывает соответствующий обработчик
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) {
	switch msg.Command() {
	case "start":
		HandleStartCommand(bot, msg)
	case "help":
		HandleHelpCommand(bot, msg)
	case "refresh_me": // ← НОВАЯ КОМАНДА
		HandleRefreshMeCommand(bot, msg, db)
		// Можно добавить другие команды
	}
}
