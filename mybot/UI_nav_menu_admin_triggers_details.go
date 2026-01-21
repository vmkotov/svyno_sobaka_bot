// ============================================================================
// ФАЙЛ: UI_nav_menu_admin_triggers_details.go
// Обработка admin:trigger:detail:TECH_KEY
// ============================================================================
package mybot

import (
	"database/sql"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleAdminTriggerDetailCallback - обработка admin:trigger:detail:TECH_KEY
func HandleAdminTriggerDetailCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	if len(parts) < 4 {
		log.Printf("⚠️ Неполный callback_data для деталей триггера: %v", parts)
		return
	}

	// Получаем триггер
	techKey := parts[3]
	trigger := GetTriggerByTechKey(techKey)

	if trigger == nil {
		log.Printf("❌ Триггер с ключом %s не найден", techKey)
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Триггер не найден")
		bot.Request(callback)
		return
	}

	log.Printf("👑 Админская детальная карточка триггера %s от @%s",
		techKey, callbackQuery.From.UserName)

	// Извлекаем номер страницы
	fromPage := extractPageFromMessage(callbackQuery.Message.Text)

	// Генерируем админскую детальную карточку
	message, keyboard := GenerateAdminTriggerDetailCard(trigger, fromPage)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		message,
		keyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки админской детальной карточки: %v", err)
	}
}
