// ============================================================================
// ФАЙЛ: UI_nav_menu_admin.go
// Обработка UI callback админки (admin:*)
// ============================================================================
package mybot

import (
	"database/sql"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleAdminUICallback - обработка UI callback админки
func HandleAdminUICallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	if len(parts) < 2 {
		log.Printf("⚠️ Неполный admin callback_data: %v", parts)
		return
	}

	switch parts[1] {
	case "menu":
		log.Printf("👑 Админское меню от @%s", callbackQuery.From.UserName)
		showAdminMenu(bot, callbackQuery)
	case "refresh":
		log.Printf("👑 Админское обновление триггеров от @%s", callbackQuery.From.UserName)
		handleAdminRefreshTriggers(bot, callbackQuery, db)
	case "triggers":
		handleAdminTriggersUICallback(bot, callbackQuery, parts, db)
	case "trigger":
		// admin:trigger:detail:TECH_KEY
		HandleAdminTriggerDetailCallback(bot, callbackQuery, parts, db)
	default:
		log.Printf("⚠️ Неизвестный admin callback: %s", parts[1])
	}
}

// showAdminMenu показывает админское меню (после нажатия на СВИНОАДМИНКА)
func showAdminMenu(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	text := "🐷 *СвиноАдминка*\n\n" +
		"Выберите действие:"

	// Создаем inline-клавиатуру с двумя кнопками
	refreshButton := tgbotapi.NewInlineKeyboardButtonData(
		"🔄 Обновить триггеры", 
		"admin:refresh",
	)
	triggersButton := tgbotapi.NewInlineKeyboardButtonData(
		"📋 Просмотр триггеров", 
		"admin:triggers:list",
	)

	// Две кнопки в один ряд
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(refreshButton, triggersButton),
	)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		text,
		inlineKeyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки админского меню: %v", err)
	}
}

// handleAdminRefreshTriggers - обновление триггеров из админки
func handleAdminRefreshTriggers(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Проверяем, что это личный чат
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Админский callback из группы, игнорируем: chat_id=%d",
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
	log.Printf("✅ Триггеры обновлены через админку от @%s", callbackQuery.From.UserName)
}

// handleAdminTriggersUICallback - обработка админских триггеров
func handleAdminTriggersUICallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	if len(parts) < 3 {
		log.Printf("⚠️ Неполный admin triggers callback: %v", parts)
		return
	}

	switch parts[2] {
	case "list":
		// Показать первую страницу админских триггеров
		log.Printf("👑 Админский список триггеров от @%s", callbackQuery.From.UserName)
		showAdminTriggersMenu(bot, callbackQuery, db)
	default:
		log.Printf("⚠️ Неизвестный admin triggers команда: %s", parts[2])
	}
}

// showAdminTriggersMenu показывает админское меню триггеров
func showAdminTriggersMenu(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Проверяем, что это личный чат
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Админский callback из группы, игнорируем: chat_id=%d",
			callbackQuery.Message.Chat.ID)
		return
	}

	// Генерируем меню первой страницы с админской навигацией
	menuText, menuKeyboard := GenerateAdminTriggersMenu(0)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		menuText,
		menuKeyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки админского меню триггеров: %v", err)
	}

	log.Printf("✅ Админское меню триггеров отправлено для @%s", callbackQuery.From.UserName)
}
