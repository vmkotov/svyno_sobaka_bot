package mybot

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"svyno_sobaka_bot/mybot/ui"  // Импортируем UI пакет
)

// handleCommand - определяет команду и вызывает соответствующий обработчик
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) {
	switch msg.Command() {
	case "start":
		ui.HandleStartCommand(bot, msg)
	case "help":
		ui.HandleHelpCommand(bot, msg)
	case "refresh_me": // ← НОВАЯ КОМАНДА
		handleRefreshMeCommand(bot, msg, db)
		// Можно добавить другие команды
	}
}

// handleRefreshMeCommand - перезагружает триггеры из БД и показывает список
func handleRefreshMeCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) {
	log.Printf("🔄 Команда /refresh_me от @%s", msg.From.UserName)

	// Проверяем подключение к БД
	if db == nil {
		log.Println("⚠️ БД не подключена, не могу обновить триггеры")
		ui.SendMessage(bot, msg.Chat.ID, "❌ БД не подключена", "ошибка")
		return
	}

	// Загружаем конфигурацию
	if err := LoadTriggerConfig(db); err != nil {
		log.Printf("❌ Ошибка загрузки триггеров: %v", err)
		ui.SendMessage(bot, msg.Chat.ID, "❌ Ошибка обновления триггеров", "ошибка")
		return
	}

	// Получаем загруженную конфигурацию
	config := GetTriggerConfig()
	if config == nil || len(config) == 0 {
		log.Println("⚠️ Конфигурация триггеров пуста после загрузки")
		ui.SendMessage(bot, msg.Chat.ID, "✅ Триггеры обновлены!\n⚠️ Но список пуст", "refresh_me")
		return
	}

	log.Println("✅ Триггеры перезагружены из БД")

	// 1. Отправляем сообщение об успехе
	ui.SendMessage(bot, msg.Chat.ID, "✅ Триггеры обновлены!", "refresh_me")

	// 2. Формируем статистику и список
	statsText := formatTriggerStats(config)
	listText := formatTriggersList(config)

	// 3. Отправляем статистику
	ui.SendMessage(bot, msg.Chat.ID, statsText, "статистика триггеров")

	// 4. Отправляем список (разбиваем если длинный)
	maxMsgLength := 4000 // Оставляем запас от 4096
	listParts := splitLongMessage(listText, maxMsgLength)

	for i, part := range listParts {
		context := "список триггеров"
		if len(listParts) > 1 {
			context = fmt.Sprintf("список триггеров (часть %d/%d)", i+1, len(listParts))
		}
		ui.SendMessage(bot, msg.Chat.ID, part, context)
	}
}
