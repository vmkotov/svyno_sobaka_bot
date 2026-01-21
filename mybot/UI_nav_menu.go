// Файл: mybot/UI_nav_menu.go
package mybot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleMenuUICallback - обработка UI callback меню
func HandleMenuUICallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	if len(parts) < 2 {
		log.Printf("⚠️ Неполный callback_data для меню: %v", parts)
		return
	}

	switch parts[1] {
	case "main":
		log.Printf("🏠 Показать главное меню для @%s", callbackQuery.From.UserName)
		// Теперь используем редактирование вместо новой отправки
		if isAdmin(callbackQuery.From.ID) {
			editAdminMenu(bot, callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
		} else {
			editUserMenu(bot, callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
		}
	case "about":
		log.Printf("❓ О боте для @%s", callbackQuery.From.UserName)
		HandleMenuAboutCallback(bot, callbackQuery)
	default:
		log.Printf("⚠️ Неизвестный тип меню: %s", parts[1])
	}
}
