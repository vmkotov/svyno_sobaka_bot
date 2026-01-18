package mybot

import (
    "log"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ================= UI ФУНКЦИИ =================

// SendMessage - отправка простых сообщений
func SendMessage(bot *tgbotapi.BotAPI, chatID int64, text, context string) {
    reply := tgbotapi.NewMessage(chatID, text)

    if _, err := bot.Send(reply); err != nil {
        log.Printf("❌ Ошибка отправки %s: %v", context, err)
    } else {
        log.Printf("✅ Отправлен %s", context)
    }
}

// escapeMarkdown - экранирование для Telegram Markdown
func escapeMarkdown(text string) string {
    specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
    result := text
    for _, char := range specialChars {
        result = strings.ReplaceAll(result, char, "\\"+char)
    }
    return result
}

// escapeHTML - экранирование для Telegram HTML (если понадобится)
func escapeHTML(text string) string {
    replacer := strings.NewReplacer(
        "&", "&amp;",
        "<", "&lt;", 
        ">", "&gt;",
    )
    return replacer.Replace(text)
}
