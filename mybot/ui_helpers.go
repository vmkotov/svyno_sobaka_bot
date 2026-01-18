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

// escapeMarkdown - экранирование для Telegram MarkdownV2
// Специальные символы для Telegram MarkdownV2:
// _ * [ ] ( ) ~ ` > # + = | { } ! \
// НО: дефис (-), точка (.) не нужно экранировать!
func escapeMarkdown(text string) string {
    if text == "" {
        return ""
    }
    
    result := text
    
    // 1. Экранируем обратные слеши
    result = strings.ReplaceAll(result, "\\", "\\\\")
    
    // 2. Экранируем специальные символы
    specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "=", "|", "{", "}", "!", "\\"}
    
    for _, char := range specialChars {
        result = strings.ReplaceAll(result, char, "\\"+char)
    }
    
    return result
}

// safeMarkdown - безопасное создание Markdown текста
func safeMarkdown(text string) string {
    // Для обычного текста используем escapeMarkdown
    return escapeMarkdown(text)
}

// safeCode - для кодовых блоков (внутри `)
func safeCode(text string) string {
    // Внутри кодовых блоков экранируем только обратные кавычки
    result := strings.ReplaceAll(text, "`", "'") // Заменяем на апостроф
    result = strings.ReplaceAll(result, "\\", "\\\\")
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
