// ============================================================================
// ФАЙЛ: ui_callbacks_refresh.go
// Обработка UI callback обновлений (refresh:*)
// ============================================================================
package mybot

import (
	"database/sql"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleRefreshUICallback - обработка UI callback обновлений
func HandleRefreshUICallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	if len(parts) < 2 {
		log.Printf("⚠️ Неполный callback_data для обновления: %v", parts)
		return
	}

	switch parts[1] {
	case "triggers":
		handleRefreshTriggersUICallback(bot, callbackQuery, db)
	default:
		log.Printf("⚠️ Неизвестный тип обновления: %s", parts[1])
	}
}

// handleRefreshTriggersUICallback обрабатывает обновление триггеров
func handleRefreshTriggersUICallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	log.Printf("🔄 Нажата кнопка обновления триггеров от @%s",
		callbackQuery.From.UserName)

	// Проверяем, что это личный чат
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Callback из группы, игнорируем: chat_id=%d",
			callbackQuery.Message.Chat.ID)
		return
	}

	// Вызываем существующую логику через виртуальное сообщение
	virtualMsg := &tgbotapi.Message{
		MessageID: callbackQuery.Message.MessageID,
		From:      callbackQuery.From,
		Chat:      callbackQuery.Message.Chat,
		Text:      "/refresh_me",
		Date:      callbackQuery.Message.Date,
	}

	HandleRefreshMeCommand(bot, virtualMsg, db)

	log.Printf("✅ Триггеры обновлены для @%s", callbackQuery.From.UserName)
}
