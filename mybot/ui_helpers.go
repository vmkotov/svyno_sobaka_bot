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
    if text == "" {
        return ""
    }
    
    specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
    result := text
    
    // Сначала экранируем обратные слеши
    result = strings.ReplaceAll(result, "\\", "\\\\")
    
    // Затем экранируем все специальные символы
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

// sendWithMarkdownFallback - отправка с fallback на обычный текст
func sendWithMarkdownFallback(bot *tgbotapi.BotAPI, chatID int64, text string) error {
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = "Markdown"
    
    if _, err := bot.Send(msg); err != nil {
        log.Printf("⚠️ Markdown ошибка: %v, пробую без Markdown", err)
        msg.ParseMode = ""
        _, err = bot.Send(msg)
        return err
    }
    
    return nil
}
