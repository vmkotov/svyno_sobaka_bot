package mybot

import (
	"database/sql"
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
