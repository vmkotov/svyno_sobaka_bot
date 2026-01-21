// ============================================================================
// ФАЙЛ: UI_nav_menu_admin_home.go
// Обработка кнопки "Главная" в админке
// ============================================================================
package mybot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleAdminHomeCallback - обработка admin:home (или menu:main из админки)
// Показывает обычное главное меню независимо от прав
func HandleAdminHomeCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	log.Printf("🏠 Главная из админки от @%s", callbackQuery.From.UserName)
	
	// ВСЕГДА показываем обычное меню при нажатии "Главная"
	EditUserMenu(bot, callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
}
