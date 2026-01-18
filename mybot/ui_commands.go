package mybot

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStartCommand - обработка команды /start
func HandleStartCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	log.Printf("🚀 Команда /start от @%s", msg.From.UserName)
	SendMainMenu(bot, msg.Chat.ID)
}

// HandleHelpCommand - обработка команды /help
func HandleHelpCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "📋 Команды:\n" +
		"/start - Начать\n" +
		"/help - Помощь\n" +
		"/refresh_me - Обновить триггеры из БД"
	SendMessage(bot, msg.Chat.ID, text, "справка")
}

// HandleRefreshMeCommand - обработка команды /refresh_me
func HandleRefreshMeCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) {
	log.Printf("🔄 Команда /refresh_me от @%s", msg.From.UserName)

	// Проверяем подключение к БД
	if db == nil {
		log.Println("⚠️ БД не подключена, не могу обновить триггеры")
		SendMessage(bot, msg.Chat.ID, "❌ БД не подключена", "ошибка")
		return
	}

	// Загружаем конфигурацию
	if err := LoadTriggerConfig(db); err != nil {
		log.Printf("❌ Ошибка загрузки триггеров: %v", err)
		SendMessage(bot, msg.Chat.ID, "❌ Ошибка обновления триггеров", "ошибка")
		return
	}

	// Получаем загруженную конфигурацию
	config := GetTriggerConfig()
	if config == nil || len(config) == 0 {
		log.Println("⚠️ Конфигурация триггеров пуста после загрузки")
		SendMessage(bot, msg.Chat.ID, "✅ Триггеры обновлены!\n⚠️ Но список пуст", "refresh_me")
		return
	}

	log.Println("✅ Триггеры перезагружены из БД")

	// 1. Отправляем сообщение об успехе
	SendMessage(bot, msg.Chat.ID, "✅ Триггеры обновлены!", "refresh_me")

	// 2. Формируем статистику и список
	statsText := FormatTriggerStats(config)
	listText := FormatTriggersList(config)

	// 3. Отправляем статистику
	SendMessage(bot, msg.Chat.ID, statsText, "статистика триггеров")

	// 4. Отправляем список (разбиваем если длинный)
	maxMsgLength := 4000 // Оставляем запас от 4096
	listParts := SplitLongMessage(listText, maxMsgLength)

	for i, part := range listParts {
		context := "список триггеров"
		if len(listParts) > 1 {
			context = fmt.Sprintf("список триггеров (часть %d/%d)", i+1, len(listParts))
		}
		SendMessage(bot, msg.Chat.ID, part, context)
	}
}
