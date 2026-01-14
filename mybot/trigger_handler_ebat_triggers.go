package mybot

import (
    "fmt"
    "log"
    "strings"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CheckEbatTriggers проверяет сообщение на наличие фраз "ебать уфу" или "ебать спартак"
// Приоритет: 1-й (самый высокий)
// Ответ: "+"
func CheckEbatTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Нормализуем текст: нижний регистр + удаляем знаки препинания
    text := normalizeText(msg.Text)
    
    // Проверяем обе фразы
    hasEbatUfu := strings.Contains(text, "ебать уфу")
    hasEbatSpartak := strings.Contains(text, "ебать спартак")
    
    if !hasEbatUfu && !hasEbatSpartak {
        return false
    }
    
    // Определяем какую фразу нашли (для логов)
    foundPhrase := ""
    if hasEbatUfu && hasEbatSpartak {
        foundPhrase = "ебать уфу и ебать спартак"
    } else if hasEbatUfu {
        foundPhrase = "ебать уфу"
    } else {
        foundPhrase = "ебать спартак"
    }
    
    log.Printf("🎯 Триггер EbatTriggers: найдено '%s' от @%s", 
               foundPhrase, msg.From.UserName)
    
    // Отправляем ответ "+"
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "+")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки '+': %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ '+'")
    
    // Логируем (используем существующую функцию логирования)
    sendEbatTriggerLogToChat(bot, msg, foundPhrase, logChatID)
    
    return true
}

// normalizeText приводит текст к нижнему регистру и удаляет знаки препинания
func normalizeText(text string) string {
    // 1. К нижнему регистру
    text = strings.ToLower(text)
    
    // 2. Удаляем знаки препинания: ,.!?- (и множественные пробелы)
    replacer := strings.NewReplacer(
        ",", " ",
        ".", " ",
        "!", " ",
        "?", " ",
        "-", " ",
        "  ", " ", // двойные пробелы -> одинарные
    )
    
    text = replacer.Replace(text)
    
    // 3. Убираем лишние пробелы
    text = strings.TrimSpace(text)
    
    return text
}

// sendEbatTriggerLogToChat логирует срабатывание триггера
func sendEbatTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                             foundPhrase string, logChatID int64) {
    
    logText := fmt.Sprintf(
        "🔔 *Триггер EbatTriggers*\n\n" +
        "✅ *Отреагировал*\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденная фраза:* `%s`\n" +
        "💬 *Ответ:* %s",
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        foundPhrase,
        "+",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога EbatTriggers: %v", err)
    }
}
