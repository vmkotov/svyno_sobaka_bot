// ============================================================================
// ФАЙЛ: UI_nav_menu_about.go
// Обработка menu:about - информация о боте
// ============================================================================
package mybot

import (
	"fmt"
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

	// Получаем статистику триггеров
	config := GetTriggerConfig()
	triggerCount := 0
	if config != nil {
		triggerCount = len(config)
	}

	// Безопасный текст без Markdown проблем
	text := fmt.Sprintf(
		"🤖 *О боте\\-свинособаке*\n\n"+
			"Я автоматически реагирую на сообщения\n"+
			"в чатах по заданным триггерам\\.\n\n"+
			"📊 Триггеров загружено: %d\n"+
			"🔄 Используйте /refresh_me чтобы обновить\n"+
			"триггеры из базы данных\\.\n\n"+
			"🐷 Свинособака \\- это состояние души\\!",
		triggerCount,
	)

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
	msg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки информации о боте: %v", err)
		// Пробуем без Markdown
		msg.ParseMode = ""
		if _, err2 := bot.Send(msg); err2 != nil {
			log.Printf("❌ Ошибка даже без Markdown: %v", err2)
		}
	}
}
