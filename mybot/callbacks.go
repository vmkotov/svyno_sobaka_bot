package mybot

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCallbackQuery - обрабатывает callback-запросы от inline-кнопок
func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	log.Printf("🔄 Callback запрос от @%s (data: %s)", 
		callbackQuery.From.UserName, callbackQuery.Data)
	
	// ===============================================
	// ПАРСИНГ ПО НОВОЙ СИСТЕМЕ
	// ===============================================
	parts := parseCallbackData(callbackQuery.Data)
	
	if len(parts) == 0 {
		handleLegacyCallback(bot, callbackQuery, db)
		return
	}
	
	// Роутинг по первой части (тип)
	switch parts[0] {
	case "menu":
		handleMenuCallback(bot, callbackQuery, parts)
	case "triggers":
		handleTriggersCallback(bot, callbackQuery, parts, db)
	case "trigger":
		handleSingleTriggerCallback(bot, callbackQuery, parts)
	case "refresh":
		handleRefreshCallback(bot, callbackQuery, parts, db)
	default:
		handleLegacyCallback(bot, callbackQuery, db)
	}
}

// ===============================================
// ОБРАБОТКА МЕНЮ (menu:*)
// ===============================================
func handleMenuCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}
	
	// Вторая часть - тип меню
	if len(parts) < 2 {
		log.Printf("⚠️ Неполный callback_data для меню: %v", parts)
		return
	}
	
	switch parts[1] {
	case "main":
		log.Printf("🏠 Показать главное меню для @%s", callbackQuery.From.UserName)
		editMessageToMainMenu(bot, callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
	default:
		log.Printf("⚠️ Неизвестный тип меню: %s", parts[1])
	}
}

// ===============================================
// ОБРАБОТКА ТРИГГЕРОВ (triggers:*)
// ===============================================
func handleTriggersCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}
	
	if len(parts) < 2 {
		log.Printf("⚠️ Неполный callback_data для триггеров: %v", parts)
		return
	}
	
	switch parts[1] {
	case "list":
		// Показать первую страницу триггеров
		log.Printf("📋 Показать список триггеров для @%s", callbackQuery.From.UserName)
		handleShowTriggersMenu(bot, callbackQuery, db)
	case "page":
		// Показать конкретную страницу
		if len(parts) < 3 {
			log.Printf("⚠️ Нет номера страницы: %v", parts)
			return
		}
		page, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Printf("❌ Неверный номер страницы: %s", parts[2])
			return
		}
		handleTriggersPage(bot, callbackQuery, page)
	default:
		log.Printf("⚠️ Неизвестная команда триггеров: %s", parts[1])
	}
}

// handleShowTriggersMenu показывает меню триггеров (первая страница)
func handleShowTriggersMenu(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Проверяем, что это личный чат
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Callback из группы, игнорируем: chat_id=%d", 
			callbackQuery.Message.Chat.ID)
		return
	}

	// Генерируем меню первой страницы
	menuText, menuKeyboard := generateTriggersMenu(0)

	// Отправляем меню
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, menuText)
	msg.ReplyMarkup = menuKeyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки меню триггеров: %v", err)
		return
	}

	log.Printf("✅ Меню триггеров отправлено для @%s", callbackQuery.From.UserName)
}

// handleTriggersPage обрабатывает переход по страницам триггеров
func handleTriggersPage(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, page int) {
	log.Printf("📋 Показать страницу триггеров %d для @%s", 
		page, callbackQuery.From.UserName)
	
	// Генерируем меню для запрошенной страницы
	menuText, menuKeyboard := generateTriggersMenu(page)
	
	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		menuText,
		menuKeyboard,
	)
	
	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка редактирования меню триггеров: %v", err)
		// Если не удалось отредактировать, отправляем новое
		newMsg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, menuText)
		newMsg.ReplyMarkup = menuKeyboard
		bot.Send(newMsg)
	}
	
	log.Printf("✅ Меню триггеров (страница %d) отправлено для @%s", 
		page, callbackQuery.From.UserName)
}

// ===============================================
// ОБРАБОТКА ОДНОГО ТРИГГЕРА (trigger:*)
// ===============================================
func handleSingleTriggerCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string) {
	// Убираем "часики" - заглушка для первой фазы
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}
	
	if len(parts) < 3 {
		log.Printf("⚠️ Неполный callback_data для триггера: %v", parts)
		return
	}
	
	switch parts[1] {
	case "detail":
		techKey := parts[2]
		log.Printf("🎯 Нажата кнопка триггера %s от @%s", 
			techKey, callbackQuery.From.UserName)
		// В будущем здесь будет детальная информация
	default:
		log.Printf("⚠️ Неизвестная команда триггера: %s", parts[1])
	}
}

// ===============================================
// ОБНОВЛЕНИЕ (refresh:*)
// ===============================================
func handleRefreshCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	if len(parts) < 2 {
		log.Printf("⚠️ Неполный callback_data для обновления: %v", parts)
		return
	}
	
	switch parts[1] {
	case "triggers":
		handleRefreshTriggersCallback(bot, callbackQuery, db)
	default:
		log.Printf("⚠️ Неизвестный тип обновления: %s", parts[1])
	}
}

// handleRefreshTriggersCallback обрабатывает обновление триггеров
func handleRefreshTriggersCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	log.Printf("🔄 Нажата кнопка обновления триггеров от @%s", 
		callbackQuery.From.UserName)

	// Проверяем, что это личный чат
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Callback из группы, игнорируем: chat_id=%d", 
			callbackQuery.Message.Chat.ID)
		return
	}

	// Вызываем существующую логику
	virtualMsg := &tgbotapi.Message{
		MessageID: callbackQuery.Message.MessageID,
		From:      callbackQuery.From,
		Chat:      callbackQuery.Message.Chat,
		Text:      "/refresh_me",
		Date:      callbackQuery.Message.Date,
	}

	handleRefreshMeCommand(bot, virtualMsg, db)

	log.Printf("✅ Триггеры обновлены для @%s", callbackQuery.From.UserName)
}

// ===============================================
// СТАРАЯ СИСТЕМА ДЛЯ ОБРАТНОЙ СОВМЕСТИМОСТИ
// ===============================================
func handleLegacyCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Старый формат callback_data без префикса
	switch callbackQuery.Data {
	case "refresh_triggers":
		// Конвертируем в новый формат
		parts := []string{"refresh", "triggers"}
		handleRefreshCallback(bot, callbackQuery, parts, db)
	case "show_triggers":
		// Конвертируем в новый формат
		parts := []string{"triggers", "list"}
		handleTriggersCallback(bot, callbackQuery, parts, db)
	default:
		log.Printf("⚠️ Неизвестный callback_data (legacy): %s", callbackQuery.Data)
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Неизвестная команда")
		bot.Request(callback)
	}
}

// parseCallbackData парсит callback_data по новой системе
func parseCallbackData(data string) []string {
	// Формат: "тип:подтип:параметр" или "тип:подтип"
	return strings.Split(data, ":")
}
