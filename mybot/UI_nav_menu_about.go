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

	// Текст о боте
	text := "🤖 *Бот-свинособака*\n\n" +
		"Я реагирую на ключевые слова в чатах.\n" +
		"Админы могут управлять триггерами через СвиноАдминку."

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
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки информации о боте: %v", err)
	}
}

// EditUserMenu - редактирует сообщение на пользовательское меню (экспортируемая)
func EditUserMenu(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	text := "Привет! Я бот-свинособака 🐷🐶\n" +
		"Я реагирую на сообщения в чатах.\n\n" +
		"Используйте /help для списка команд."

	// Кнопки
	aboutButton := tgbotapi.NewInlineKeyboardButtonData("❓ О боте", "menu:about")
	adminButton := tgbotapi.NewInlineKeyboardButtonData("🐷 СвиноАдминка", "admin:menu")

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(aboutButton, adminButton),
	)

	// РЕДАКТИРУЕМ существующее сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		chatID,
		messageID,
		text,
		inlineKeyboard,
	)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка редактирования пользовательского меню: %v", err)
	}
}

// EditAdminMenu - редактирует сообщение на админское меню (экспортируемая)
func EditAdminMenu(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	text := "🐷 *СвиноАдминка*\n\nВыберите действие:"

	// Кнопки
	refreshButton := tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить", "admin:refresh")
	triggersButton := tgbotapi.NewInlineKeyboardButtonData("📋 Триггеры", "admin:triggers:list")
	homeButton := tgbotapi.NewInlineKeyboardButtonData("🏠 Домой", "menu:main")

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(refreshButton, triggersButton, homeButton),
	)

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		chatID,
		messageID,
		text,
		inlineKeyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка редактирования админского меню: %v", err)
	}
}
