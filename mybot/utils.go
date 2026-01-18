package mybot

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendMessage - общая функция отправки сообщений
// Может использоваться любым модулем (broadcast, commands и т.д.)
func SendMessage(bot *tgbotapi.BotAPI, chatID int64, text, context string) {
	reply := tgbotapi.NewMessage(chatID, text)

	if _, err := bot.Send(reply); err != nil {
		log.Printf("❌ Ошибка отправки %s: %v", context, err)
	} else {
		log.Printf("✅ Отправлен %s", context)
	}
}

// escapeMarkdown экранирует символы Markdown (общая функция)
func escapeMarkdown(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}
