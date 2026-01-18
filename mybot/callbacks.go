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
	
	// ===============================================
	// МАРШРУТИЗАЦИЯ ПО TИПУ CALLBACK
	// ===============================================
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

// ===============================================
// ОБНОВЛЕНИЕ ТРИГГЕРОВ
// ===============================================
func handleRefreshCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Убираем "часики" в клиенте Telegram
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
// ПОКАЗ СПИСКА ТРИГГЕРОВ
// ===============================================
func handleShowTriggersCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Убираем "часики" в клиенте Telegram
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
	// 2. ПОЛУЧЕНИЕ КОНФИГУРАЦИИ
	// ===============================================
	config := GetTriggerConfig()
	if config == nil || len(config) == 0 {
		log.Println("⚠️ Конфигурация триггеров пуста")
		sendMessage(bot, callbackQuery.Message.Chat.ID, 
			"❌ Триггеры не загружены\nИспользуйте кнопку \"🔄 Обновить триггеры\"", 
			"ошибка показа триггеров")
		return
	}

	log.Printf("📊 Показываю %d триггеров для @%s", len(config), callbackQuery.From.UserName)

	// ===============================================
	// 3. ФОРМИРОВАНИЕ СПИСКА
	// ===============================================
	listText := formatTriggersList(config)

	// ===============================================
	// 4. ОТПРАВКА СПИСКА (РАЗБИВАЕМ ЕСЛИ ДЛИННЫЙ)
	// ===============================================
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
