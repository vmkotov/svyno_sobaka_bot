// ============================================================================
// ФАЙЛ: ui_callbacks.go
// Главный роутер UI callback-запросов
// Делегирует обработку специализированным UI модулям
// ============================================================================
package mybot

import (
	"database/sql"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCallbackQuery - главный обработчик UI callback-запросов
func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	log.Printf("🔄 Callback запрос от @%s (data: %s, ID: %d)",
		callbackQuery.From.UserName, callbackQuery.Data, callbackQuery.From.ID)

	// Проверка доступа для админских callback
	if strings.HasPrefix(callbackQuery.Data, "admin:") {
		if !checkAdminAccess(callbackQuery.From.ID, callbackQuery.Data) {
			log.Printf("🚫 Отказ в доступе: @%s пытался использовать админский callback",
				callbackQuery.From.UserName)
			
			callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Ты свинособака, а не ОДМИН! 🐷")
			bot.Request(callback)
			return
		}
		log.Printf("👑 Админский доступ разрешен для @%s", 
			callbackQuery.From.UserName)
	}

	// Парсинг callback_data
	parts := parseCallbackData(callbackQuery.Data)

	if len(parts) == 0 {
		handleLegacyCallback(bot, callbackQuery, db)
		return
	}

	// Роутинг по типу callback
	switch parts[0] {
	case "menu":
		HandleMenuUICallback(bot, callbackQuery, parts)
	case "refresh":
		HandleRefreshUICallback(bot, callbackQuery, parts, db)
	case "admin":
		HandleAdminUICallback(bot, callbackQuery, parts, db)
	default:
		handleLegacyCallback(bot, callbackQuery, db)
	}
}

// parseCallbackData парсит callback_data по системе "тип:подтип:параметр"
func parseCallbackData(data string) []string {
	return strings.Split(data, ":")
}

// handleLegacyCallback - обработка старых форматов callback_data
func handleLegacyCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Старый формат callback_data без префикса
	switch callbackQuery.Data {
	case "refresh_triggers":
		// Конвертируем в новый формат
		parts := []string{"refresh", "triggers"}
		HandleRefreshUICallback(bot, callbackQuery, parts, db)
	case "show_triggers":
		// Конвертируем в новый формат
		parts := []string{"admin", "triggers", "list"}
		HandleAdminUICallback(bot, callbackQuery, parts, db)
	default:
		log.Printf("⚠️ Неизвестный callback_data (legacy): %s", callbackQuery.Data)
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Неизвестная команда")
		bot.Request(callback)
	}
}
