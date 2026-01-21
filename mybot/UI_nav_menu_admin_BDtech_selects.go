package mybot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBDtechSelectsCallback - заглушка для раздела SELECT запросов
func HandleBDtechSelectsCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string) {
	text := "🔍 *БД Тех - SELECT запросы*\n\nРаздел в разработке 🛠️\n\nСкоро здесь будет:\n• Выполнение запросов\n• Фильтрация данных\n• Пагинация результатов"

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
		log.Printf("❌ Ошибка отправки заглушки SELECT: %v", err)
	}
}
