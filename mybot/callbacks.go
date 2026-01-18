package mybot

import (
	"database/sql"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCallbackQuery - обрабатывает callback-запросы от inline-кнопок
func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	log.Printf("🔄 Callback запрос от @%s (data: %s)", 
		callbackQuery.From.UserName, callbackQuery.Data)
	
	// ===============================================
	// ПАРСИНГ И МАРШРУТИЗАЦИЯ CALLBACK_DATA
	// ===============================================
	callbackType, callbackValue := parseCallbackData(callbackQuery.Data)
	
	switch callbackType {
	case "refresh_triggers":
		handleRefreshCallback(bot, callbackQuery, db)
	case "show_triggers":
		handleShowTriggersCallback(bot, callbackQuery, db)
	case "triggers_page":
		handleTriggersPageCallback(bot, callbackQuery, callbackValue)
	case "trigger_info":
		handleTriggerInfoCallback(bot, callbackQuery, callbackValue)
	default:
		// Старая система для обратной совместимости
		handleLegacyCallback(bot, callbackQuery, db)
	}
}

// ===============================================
// ОБНОВЛЕНИЕ ТРИГГЕРОВ (СТАРАЯ СИСТЕМА)
// ===============================================
func handleRefreshCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	log.Printf("🔄 Нажата кнопка обновления триггеров от @%s", 
		callbackQuery.From.UserName)

	// ===============================================
	// 1. ПРОВЕРКА: ТОЛЬКО ЛИЧНЫЙ ЧАТ
	// ===============================================
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Callback из группы, игнорируем: chat_id=%d", 
			callbackQuery.Message.Chat.ID)
		return
	}

	// ===============================================
	// 2. ВЫЗОВ СУЩЕСТВУЮЩЕЙ ЛОГИКИ
	// ===============================================
	virtualMsg := &tgbotapi.Message{
		MessageID: callbackQuery.Message.MessageID,
		From:      callbackQuery.From,
		Chat:      callbackQuery.Message.Chat,
		Text:      "/refresh_me",
		Date:      callbackQuery.Message.Date,
	}

	handleRefreshMeCommand(bot, virtualMsg, db)

	log.Printf("✅ Callback обработан для @%s", callbackQuery.From.UserName)
}

// ===============================================
// ПОКАЗ СПИСКА ТРИГГЕРОВ (НОВАЯ СИСТЕМА)
// ===============================================
func handleShowTriggersCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	log.Printf("📋 Нажата кнопка показа триггеров от @%s", 
		callbackQuery.From.UserName)

	// ===============================================
	// 1. ПРОВЕРКА: ТОЛЬКО ЛИЧНЫЙ ЧАТ
	// ===============================================
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Callback из группы, игнорируем: chat_id=%d", 
			callbackQuery.Message.Chat.ID)
		return
	}

	// ===============================================
	// 2. ГЕНЕРАЦИЯ МЕНЮ ПЕРВОЙ СТРАНИЦЫ
	// ===============================================
	menuText, menuKeyboard := generateTriggersMenu(0)

	// ===============================================
	// 3. ОТПРАВКА МЕНЮ
	// ===============================================
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, menuText)
	msg.ReplyMarkup = menuKeyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки меню триггеров: %v", err)
		return
	}

	log.Printf("✅ Меню триггеров отправлено для @%s", callbackQuery.From.UserName)
}

// ===============================================
// ОБРАБОТКА СТРАНИЦ ТРИГГЕРОВ
// ===============================================
func handleTriggersPageCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, pageStr string) {
	// Парсим номер страницы
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		log.Printf("❌ Неверный номер страницы: %s", pageStr)
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Ошибка")
		bot.Request(callback)
		return
	}
	
	handleTriggerPageCallback(bot, callbackQuery, page)
}

// handleTriggerPageCallback обрабатывает переход по страницам триггеров
func handleTriggerPageCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, page int) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}
	
	log.Printf("📋 Показать страницу триггеров %d для @%s", 
		page, callbackQuery.From.UserName)
	
	// Генерируем меню для запрошенной страницы
	menuText, menuKeyboard := generateTriggersMenu(page)
	
	// Отправляем/редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		menuText,
		menuKeyboard,
	)
	
	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки меню триггеров: %v", err)
		// Если не удалось отредактировать, отправляем новое
		newMsg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, menuText)
		newMsg.ReplyMarkup = menuKeyboard
		bot.Send(newMsg)
	}
	
	log.Printf("✅ Меню триггеров (страница %d) отправлено для @%s", 
		page, callbackQuery.From.UserName)
}

// ===============================================
// ОБРАБОТКА НАЖАТИЯ НА ТРИГГЕР
// ===============================================
func handleTriggerInfoCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, techKey string) {
	// Просто убираем "часики" - заглушка для первой фазы
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}
	
	log.Printf("🎯 Нажата кнопка триггера %s от @%s", 
		techKey, callbackQuery.From.UserName)
	
	// В будущем здесь будет детальная информация о триггере
	// Пока просто логируем
}

// ===============================================
// СТАРАЯ СИСТЕМА ДЛЯ ОБРАТНОЙ СОВМЕСТИМОСТИ
// ===============================================
func handleLegacyCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Старый формат callback_data без префикса
	switch callbackQuery.Data {
	case "refresh_triggers":
		handleRefreshCallback(bot, callbackQuery, db)
	case "show_triggers":
		handleShowTriggersCallback(bot, callbackQuery, db)
	default:
		log.Printf("⚠️ Неизвестный callback_data: %s", callbackQuery.Data)
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Неизвестная команда")
		bot.Request(callback)
	}
}

// parseCallbackData парсит callback_data для меню триггеров
func parseCallbackData(data string) (string, string) {
	// Форматы:
	// "triggers_page:1" -> ("page", "1")
	// "trigger_info:tech_key" -> ("info", "tech_key")
	
	parts := splitCallbackData(data)
	if len(parts) != 2 {
		return "", ""
	}
	
	return parts[0], parts[1] // тип, значение
}

// splitCallbackData разделяет callback_data по первому двоеточию
func splitCallbackData(data string) []string {
	for i := 0; i < len(data); i++ {
		if data[i] == ':' {
			return []string{data[:i], data[i+1:]}
		}
	}
	return []string{}
}
