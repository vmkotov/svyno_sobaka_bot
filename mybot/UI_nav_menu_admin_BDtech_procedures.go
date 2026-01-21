package mybot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBDtechProceduresCallback - заглушка для раздела процедур
func HandleBDtechProceduresCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string) {
	text := "⚙️ *БД Тех - Хранимые процедуры*\n\nРаздел в разработке 🛠️\n\nСкоро здесь будет:\n• Список процедур\n• Информация о процедурах\n• Вызов процедур"

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
		log.Printf("❌ Ошибка отправки заглушки процедур: %v", err)
	}
}
