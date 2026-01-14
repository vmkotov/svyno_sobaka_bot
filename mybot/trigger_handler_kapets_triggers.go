package mybot

import (
    "fmt"
    "log"
    "strings"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Список триггерных слов (в нижнем регистре для проверки)
var kapetsWords = []string{
    "капец",
    "пиздец", 
    "пздц",
}

// CheckKapetsTriggers проверяет сообщение на слова капец/пиздец/пздц
// Приоритет: 6-й (самый последний)
func CheckKapetsTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Нормализуем текст
    text := normalizeText(msg.Text)
    
    // Проверяем триггерные слова
    foundWords := []string{}
    for _, word := range kapetsWords {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Если ничего не найдено
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("💥 Триггер KapetsTriggers: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    // Отправляем ответ
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "Трактор опять проебал?")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки KapetsTriggers: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ: Трактор опять проебал?")
    
    // Логируем
    sendKapetsTriggerLogToChat(bot, msg, foundWords, logChatID)
    
    return true
}

// sendKapetsTriggerLogToChat логирует срабатывание триггера
func sendKapetsTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                               foundWords []string, logChatID int64) {
    
    logText := fmt.Sprintf(
        "🔔 *Триггер KapetsTriggers*\n\n" +
        "✅ *Отреагировал*\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные слова:* %v\n" +
        "📊 *Всего слов:* %d\n" +
        "💬 *Ответ:* %s",
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        foundWords,
        len(foundWords),
        "Трактор опять проебал?",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога KapetsTriggers: %v", err)
    }
}
