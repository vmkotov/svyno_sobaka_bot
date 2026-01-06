package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// triggerWordsSvinoSobaka - список триггерных слов (в нижнем регистре)
var triggerWordsSvinoSobaka = []string{
    "пёс", "псом", "собака", "собаке", "собакам", "собаки", "собак", 
    "собачка", "собачонка", "собачник", "собачница", "собачатина", 
    "свинья", "свинье", "свиньям", "свиней", "свиньи", "свин", 
    "свинка", "свинёнок", "свинтус", "свинюшка", "свинство", 
    "свинарник", "свинарня", "свиноферма", "свиносовхоз", 
    "свинокомплекс", "свиноматка", "свиноводство", "свиновод", 
    "свинарь", "свинарка", "свинобой", "свинобоец", "свинопас", 
    "свинина", "свинокопчёности", "свинуха", "свинушка", 
    "свинобармен", "свинота", "свинотека", "дзюба",
    "собачий", "собакин", "свиной", "свинский", "свинячий", 
    "свинокопчёный", "свиноподобный", "свиноводческий",
    "собачиться", "присобачить", "свинячить", "насвинячить", 
    "насвинячиться",
    "насвиняченный", "насвинячивший",
}

// reactionProbability - вероятность реакции (0.5 = 50%)
const reactionProbability = 0.5

// CheckSvinoSobakaTriggers проверяет сообщение на триггерные слова
// и с вероятностью 50% отправляет реплай "Свинособака"
func CheckSvinoSobakaTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Приводим текст к нижнему регистру для проверки
    textLower := strings.ToLower(msg.Text)
    
    // Проверяем каждое триггерное слово
    var foundWords []string
    for _, word := range triggerWordsSvinoSobaka {
        if strings.Contains(textLower, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Если триггерные слова не найдены
    if len(foundWords) == 0 {
        return false
    }
    
    // Логируем обнаружение
    log.Printf("🔍 Найдены триггерные слова: %v", foundWords)
    
    // Проверяем вероятность реакции
    rand.Seed(time.Now().UnixNano())
    if rand.Float64() > reactionProbability {
        log.Printf("🎲 Пропускаем реакцию (не выпала вероятность 50%%)")
        
        // Но всё равно логируем в лог-чат
        sendTriggerLogToChat(bot, msg, foundWords, false, logChatID)
        return false
    }
    
    // Отправляем реплай "Свинособака"
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "Свинособака")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки реплая: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен реплай на триггерные слова: %v", foundWords)
    
    // Логируем в лог-чат
    sendTriggerLogToChat(bot, msg, foundWords, true, logChatID)
    
    return true
}

// sendTriggerLogToChat отправляет информацию о триггере в лог-чат
func sendTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                         foundWords []string, reacted bool, logChatID int64) {
    
    var reactionStatus string
    if reacted {
        reactionStatus = "✅ *Отреагировал*"
    } else {
        reactionStatus = "🎲 *Пропущено (вероятность)*"
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер Свинособака*\n\n" +
        "%s\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные слова:* %v\n" +
        "📊 *Всего слов:* %d",
        reactionStatus,
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        foundWords,
        len(foundWords),
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога триггера: %v", err)
    }
}
