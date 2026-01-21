// Файл: mybot/UI_nav_menu_about.go
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

	// Кнопка "Назад" - ТЕПЕРЬ с функцией редактирования!
	editMainMenu(bot, callbackQuery, text)
}

// editMainMenu - редактирует сообщение с главным меню
func editMainMenu(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, currentText string) {
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.MessageID

	// Определяем, какое меню показывать
	if isAdmin(callbackQuery.From.ID) {
		editAdminMenu(bot, chatID, messageID)
	} else {
		editUserMenu(bot, chatID, messageID)
	}
}

// editUserMenu - редактирует сообщение на пользовательское меню
func editUserMenu(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
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

// editAdminMenu - редактирует сообщение на админское меню
func editAdminMenu(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
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
