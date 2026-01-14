package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Списки триггерных слов (в нижнем регистре для проверки)
var alcoholNouns = []string{
    "праздник", "праздники", "праздниках", "праздником",
    "пиво", "пивко", "пивас", "пивасик", "пивасика", "пивасику",
    "пива", "пивко", "пиву", "пивку",
    "виски", "вискарь", "вискаря", "вискарём",
    "настойка", "настойки", "настойку", "настойке",
}

var alcoholAdjectives = []string{
    "ячменное", "ячменного", 
    "светлое", "светлого", 
    "тёмное", "тёмного", 
    "игристое",
}

var alcoholVerbs = []string{
    "ёбнуть", "йобнем", 
    "выпьем", "выпить", "выпили", 
    "нахуяримся", "бахнем", "нахуяриться",
}

var alcoholPhrases = []string{
    "день рождения", 
    "с днём рождения",
}

// Варианты ответов
var alcoholResponses = []string{
    "Давайте выпьем!",
    "Давай йобнем!",
    "Когда в бар пойдём?",
}

// CheckAlcoholTriggers проверяет сообщение на алкогольные триггеры
// Приоритет: 5-й (последний)
func CheckAlcoholTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Нормализуем текст (используем уже созданную функцию normalizeText)
    text := normalizeText(msg.Text)
    
    // Проверяем все триггерные слова и фразы
    foundWords := []string{}
    
    // Проверяем существительные
    for _, word := range alcoholNouns {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Проверяем прилагательные
    for _, word := range alcoholAdjectives {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Проверяем глаголы
    for _, word := range alcoholVerbs {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Проверяем словосочетания
    for _, phrase := range alcoholPhrases {
        if strings.Contains(text, phrase) {
            foundWords = append(foundWords, phrase)
        }
    }
    
    // Если ничего не найдено
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("🍺 Триггер AlcoholTriggers: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    // Выбираем случайный ответ
    rand.Seed(time.Now().UnixNano())
    responseIndex := rand.Intn(len(alcoholResponses))
    response := alcoholResponses[responseIndex]
    
    // Отправляем ответ
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки алкогольного триггера: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен алкогольный ответ: %s", response)
    
    // Логируем
    sendAlcoholTriggerLogToChat(bot, msg, foundWords, responseIndex, logChatID)
    
    return true
}

// sendAlcoholTriggerLogToChat логирует срабатывание алкогольного триггера
func sendAlcoholTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                foundWords []string, responseIndex int, logChatID int64) {
    
    // Обрезаем список слов если их много
    wordsForLog := foundWords
    if len(foundWords) > 5 {
        wordsForLog = foundWords[:5]
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер AlcoholTriggers*\n\n" +
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
        wordsForLog,
        len(foundWords),
        alcoholResponses[responseIndex],
        responseIndex+1,
        len(alcoholResponses),
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога AlcoholTriggers: %v", err)
    }
}
