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

// escapeMarkdown - экранирование для Telegram Markdown V2
// Специальные символы для Telegram MarkdownV2:
// _ * [ ] ( ) ~ ` > # + - = | { } . !
// НО: дефис (-) не нужно экранировать в большинстве случаев!
func escapeMarkdown(text string) string {
    if text == "" {
        return ""
    }
    
    // Специальные символы которые НУЖНО экранировать в Telegram Markdown
    // Источник: https://core.telegram.org/bots/api#markdownv2-style
    specialChars := []string{
        "_", "*", "[", "]", "(", ")", "~", "`", 
        ">", "#", "+", "=", "|", "{", "}", ".", "!",
        // Дефис (-) обычно НЕ нужно экранировать, только если он часть конструкции
    }
    
    result := text
    
    // Сначала экранируем обратные слеши
    result = strings.ReplaceAll(result, "\\", "\\\\")
    
    // Затем экранируем все специальные символы
    for _, char := range specialChars {
        result = strings.ReplaceAll(result, char, "\\"+char)
    }
    
    return result
}

// escapeMarkdownV2 - более умная версия, которая не экранирует дефисы в словах
func escapeMarkdownV2(text string) string {
    if text == "" {
        return ""
    }
    
    result := text
    
    // 1. Экранируем обратные слеши
    result = strings.ReplaceAll(result, "\\", "\\\\")
    
    // 2. Экранируем специальные символы, но осторожно с дефисами
    // Список символов которые ВСЕГДА нужно экранировать
    alwaysEscape := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "=", "|", "{", "}", "!", "\\"}
    
    for _, char := range alwaysEscape {
        result = strings.ReplaceAll(result, char, "\\"+char)
    }
    
    // 3. Точку экранируем только если она не в числе или URL
    // (упрощенная версия)
    result = strings.ReplaceAll(result, ".", "\\.")
    
    return result
}

// safeMarkdown - безопасное создание Markdown текста
func safeMarkdown(text string) string {
    // Для обычного текста (не кода) используем умную версию
    return escapeMarkdownV2(text)
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
