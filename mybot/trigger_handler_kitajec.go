package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Списки триггерных слов китаец (в нижнем регистре)
var kitajecNouns = []string{
    "китаец", "китайцы", "китайцам",
}

var kitajecAdjectives = []string{
    "китайский", "китайские", "китайская",
}

// Варианты ответов
var kitajecResponses = []string{
    "Макс китаец опять пропал?",
    "Китаец, ты где?",
}

// CheckKitajecTriggers проверяет сообщение на слова китаец/китайский
// Приоритет: 11-й (самый последний)
// Вероятность: 100% (всегда)
// Ответ: случайный из 2 вариантов (50/50)
func CheckKitajecTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Нормализуем текст
    text := normalizeText(msg.Text)
    
    // Проверяем все триггерные слова
    foundWords := []string{}
    
    // Проверяем существительные
    for _, word := range kitajecNouns {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Проверяем прилагательные
    for _, word := range kitajecAdjectives {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Если ничего не найдено
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("🇨🇳 Триггер Kitajec: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    // Выбираем случайный ответ (50/50)
    rand.Seed(time.Now().UnixNano())
    responseIndex := rand.Intn(len(kitajecResponses))
    response := kitajecResponses[responseIndex]
    
    // Отправляем ответ (всегда)
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки Kitajec: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ Kitajec: %s", response)
    
    // Логируем
    sendKitajecTriggerLogToChat(bot, msg, foundWords, responseIndex, logChatID)
    
    return true
}

// sendKitajecTriggerLogToChat логирует срабатывание триггера
func sendKitajecTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                foundWords []string, responseIndex int, logChatID int64) {
    
    logText := fmt.Sprintf(
        "🔔 *Триггер Kitajec*\n\n" +
        "✅ *Отреагировал* (всегда 100%%)\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные слова:* %v\n" +
        "📊 *Всего слов:* %d\n" +
        "💬 *Ответ:* %s\n" +
        "🔢 *Вариант ответа:* %d/%d",
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        foundWords,
        len(foundWords),
        kitajecResponses[responseIndex],
        responseIndex+1,
        len(kitajecResponses),
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога Kitajec: %v", err)
    }
}
