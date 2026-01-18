package mybot

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCallbackQuery - обрабатывает callback-запросы от inline-кнопок
func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	log.Printf("🔄 Callback запрос от @%s (data: %s)", 
		callbackQuery.From.UserName, callbackQuery.Data)
	
	// Маршрутизируем callback по его data
	switch callbackQuery.Data {
	case "refresh_triggers":
		handleRefreshCallback(bot, callbackQuery, db)
	case "show_triggers":
		handleShowTriggersCallback(bot, callbackQuery, db)
	default:
		log.Printf("⚠️ Неизвестный callback_data: %s", callbackQuery.Data)
		// Можно отправить уведомление пользователю
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Неизвестная команда")
		bot.Request(callback)
	}
}

// handleRefreshCallback - обработка нажатия кнопки обновления триггеров
func handleRefreshCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Убираем "часики" в клиенте Telegram
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	log.Printf("🔄 Нажата кнопка обновления триггеров от @%s", 
		callbackQuery.From.UserName)

	// Проверяем, что это личный чат (как договорились)
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Callback из группы, игнорируем: chat_id=%d", 
			callbackQuery.Message.Chat.ID)
		return
	}

	// Создаем виртуальное сообщение для вызова существующей логики
	virtualMsg := &tgbotapi.Message{
		MessageID: callbackQuery.Message.MessageID,
		From:      callbackQuery.From,
		Chat:      callbackQuery.Message.Chat,
		Text:      "/refresh_me",
		Date:      callbackQuery.Message.Date,
	}

	// Вызываем существующую логику обновления триггеров
	handleRefreshMeCommand(bot, virtualMsg, db)

	log.Printf("✅ Callback обработан для @%s", callbackQuery.From.UserName)
}

// handleShowTriggersCallback - обработка нажатия кнопки показа триггеров
func handleShowTriggersCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Убираем "часики" в клиенте Telegram
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	log.Printf("📋 Нажата кнопка показа триггеров от @%s", 
		callbackQuery.From.UserName)

	// Проверяем, что это личный чат
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Callback из группы, игнорируем: chat_id=%d", 
			callbackQuery.Message.Chat.ID)
		return
	}

	// Получаем текущую конфигурацию триггеров
	config := GetTriggerConfig()
	if config == nil || len(config) == 0 {
		log.Println("⚠️ Конфигурация триггеров пуста")
		sendMessage(bot, callbackQuery.Message.Chat.ID, 
			"❌ Триггеры не загружены\nИспользуйте кнопку \"🔄 Обновить триггеры\"", 
			"ошибка показа триггеров")
		return
	}

	log.Printf("📊 Показываю %d триггеров для @%s", len(config), callbackQuery.From.UserName)

	// Формируем список триггеров (только список, без статистики)
	listText := formatTriggersList(config)

	// Отправляем заголовок
	sendMessage(bot, callbackQuery.Message.Chat.ID, 
		"📋 Список по приоритету:", 
		"заголовок списка триггеров")

	// Отправляем список (разбиваем если длинный)
	maxMsgLength := 4000 // Оставляем запас от 4096
	listParts := splitLongMessage(listText, maxMsgLength)

	for i, part := range listParts {
		context := "список триггеров"
		if len(listParts) > 1 {
			context = fmt.Sprintf("список триггеров (часть %d/%d)", i+1, len(listParts))
		}
		sendMessage(bot, callbackQuery.Message.Chat.ID, part, context)
	}

	log.Printf("✅ Список триггеров отправлен для @%s", callbackQuery.From.UserName)
}
