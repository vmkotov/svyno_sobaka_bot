package mybot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBDtechJSONCallback - заглушка для раздела JSON операций
func HandleBDtechJSONCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string) {
	text := "📄 *БД Тех - JSON операции*\n\nРаздел в разработке 🛠️\n\nСкоро здесь будет:\n• Экспорт данных в JSON\n• Импорт JSON\n• JSON функции PostgreSQL"

	// Кнопка возврата
	backBtn := tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в BDtech", "admin:bdtech:menu")

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(backBtn),
	)

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		text,
		inlineKeyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки заглушки JSON: %v", err)
	}
}
