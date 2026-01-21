// ============================================================================
// ФАЙЛ: UI_nav_menu_about.go
// Обработка menu:about - информация о боте
// ============================================================================
package mybot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleMenuAboutCallback - обработка menu:about
func HandleMenuAboutCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	log.Printf("❓ О боте от @%s", callbackQuery.From.UserName)

	// Простой текст
	text := "иди нахуй собака"

	// Кнопка "Назад"
	backButton := tgbotapi.NewInlineKeyboardButtonData("🏠 Назад", "menu:main")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(backButton),
	)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		text,
		keyboard,
	)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки информации о боте: %v", err)
	}
}
