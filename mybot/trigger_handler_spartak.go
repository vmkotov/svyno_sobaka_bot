package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// spartakTriggerWords - список триггерных слов Спартак (в нижнем регистре)
var spartakTriggerWords = []string{
    "спартак", "срамтак", "спартач", "спартаку", "спартаковец", 
    "спартака", "спартачам", "спартаки", "спартаком",
}

// spartakResponses - возможные ответы на триггер Спартак
var spartakResponses = []string{
    "Ебать спартак!",
    "От Москвы и до Баку в рот давали спартаку!",
    "Пидорва!",
    `Мы выпьем вашу водку! 
Мы трахнем ваших Баб
Ебать спартак московский!
*Воистину ебать!!*`,
}

// CheckSpartakTriggers проверяет сообщение на триггерные слова Спартак
// и отправляет случайный ответ из списка
func CheckSpartakTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Приводим текст к нижнему регистру для проверки
    textLower := strings.ToLower(msg.Text)
    
    // Проверяем каждое триггерное слово
    var foundWords []string
    for _, word := range spartakTriggerWords {
        if strings.Contains(textLower, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Если триггерные слова не найдены
    if len(foundWords) == 0 {
        return false
    }
    
    // Логируем обнаружение
    log.Printf("🔍 Найдены триггерные слова Спартак: %v", foundWords)
    
    // Выбираем случайный ответ
    rand.Seed(time.Now().UnixNano())
    responseIndex := rand.Intn(len(spartakResponses))
    response := spartakResponses[responseIndex]
    
    // Отправляем реплай
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
    replyMsg.ReplyToMessageID = msg.MessageID
    replyMsg.ParseMode = "Markdown" // Для жирного текста в последнем ответе
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки реплая Спартак: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен реплай Спартак (вариант %d): %s", 
        responseIndex+1, strings.Split(response, "\n")[0])
    
    // Логируем в лог-чат
    sendSpartakTriggerLogToChat(bot, msg, foundWords, responseIndex, logChatID)
    
    return true
}

// sendSpartakTriggerLogToChat отправляет информацию о триггере Спартак в лог-чат
func sendSpartakTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                foundWords []string, responseIndex int, logChatID int64) {
    
    // Обрезаем длинный ответ для лога
    shortResponse := spartakResponses[responseIndex]
    if len(shortResponse) > 50 {
        shortResponse = shortResponse[:50] + "..."
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер Спартак*\n\n" +
        "✅ *Отреагировал*\n" +
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
        escapeMarkdown(shortResponse),
        responseIndex+1,
        len(spartakResponses),
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога триггера Спартак: %v", err)
    }
}
