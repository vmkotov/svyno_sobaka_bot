package mybot

import (
	"database/sql"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCommand - определяет команду и вызывает соответствующий обработчик
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) {
	switch msg.Command() {
	case "start":
		handleStartCommand(bot, msg)
	case "help":
		handleHelpCommand(bot, msg)
	case "refresh_me": // ← НОВАЯ КОМАНДА
		handleRefreshMeCommand(bot, msg, db)
		// Можно добавить другие команды
	}
}

// handleRefreshMeCommand - перезагружает триггеры из БД
func handleRefreshMeCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) {
	log.Printf("🔄 Команда /refresh_me от @%s", msg.From.UserName)

	// ПРОСТО: грузим триггеры
	if db == nil {
		log.Println("⚠️ БД не подключена, не могу обновить триггеры")
		sendMessage(bot, msg.Chat.ID, "❌ БД не подключена", "ошибка")
		return
	}

	// ПРОСТО: загружаем конфигурацию
	if err := LoadTriggerConfig(db); err != nil {
		log.Printf("❌ Ошибка загрузки триггеров: %v", err)
		sendMessage(bot, msg.Chat.ID, "❌ Ошибка обновления триггеров", "ошибка")
		return
	}

	// ПРОСТО: сообщаем об успехе
	sendMessage(bot, msg.Chat.ID, "✅ Триггеры обновлены!", "refresh_me")
	log.Println("✅ Триггеры перезагружены из БД")
}
