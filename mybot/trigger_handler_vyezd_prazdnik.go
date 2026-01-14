package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Список слов выезд
var vyezdWords = []string{
    "выезд", "выезде", "выездов", "выездам", "выезда",
}

// Варианты ответов
var vyezdResponses = []string{
    "Выезд – это праздник!",
    "Гонять всегда надо!",
}

// CheckVyezdPrazdnikTriggers проверяет сообщение на слово выезд
// Приоритет: 17-й (самый последний)
// Вероятность: 100% (всегда) + случайный выбор из 2 вариантов
func CheckVyezdPrazdnikTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    text := normalizeText(msg.Text)
    foundWords := []string{}
    
    for _, word := range vyezdWords {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("🚗 Триггер VyezdPrazdnik: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    // Выбираем случайный ответ (50/50)
    rand.Seed(time.Now().UnixNano())
    responseIndex := rand.Intn(len(vyezdResponses))
    response := vyezdResponses[responseIndex]
    
    // Отправляем ответ (всегда 100%)
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки VyezdPrazdnik: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ VyezdPrazdnik: вариант %d", responseIndex+1)
    
    // Логируем
    sendVyezdPrazdnikTriggerLogToChat(bot, msg, foundWords, responseIndex, logChatID)
    
    return true
}

func sendVyezdPrazdnikTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                      foundWords []string, responseIndex int, logChatID int64) {
    
    logText := fmt.Sprintf(
        "🔔 *Триггер VyezdPrazdnik*\n\n" +
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
        vyezdResponses[responseIndex],
        responseIndex+1,
        len(vyezdResponses),
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога VyezdPrazdnik: %v", err)
    }
}
