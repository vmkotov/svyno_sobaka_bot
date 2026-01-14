package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Список имён Виталик
var vitalikNames = []string{
    "виталик", "виталя", "виталь", "виталий",
}

// CheckVitalikViebatTriggers проверяет сообщение на имя Виталик
// Приоритет: 14-й
// Вероятность: 25% (каждое 4-е примерно)
func CheckVitalikViebatTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    text := normalizeText(msg.Text)
    foundNames := []string{}
    
    for _, name := range vitalikNames {
        if strings.Contains(text, name) {
            foundNames = append(foundNames, name)
        }
    }
    
    if len(foundNames) == 0 {
        return false
    }
    
    log.Printf("👤 Триггер VitalikViebat: найдено %d имён от @%s", 
               len(foundNames), msg.From.UserName)
    
    rand.Seed(time.Now().UnixNano())
    if rand.Float64() > 0.25 { // 75% шанс пропустить
        log.Printf("🎲 Пропущено VitalikViebat (вероятность 25%%)")
        sendVitalikViebatTriggerLogToChat(bot, msg, foundNames, false, logChatID)
        return false
    }
    
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "Осторожно, может въебать!")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки VitalikViebat: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ VitalikViebat")
    sendVitalikViebatTriggerLogToChat(bot, msg, foundNames, true, logChatID)
    return true
}

func sendVitalikViebatTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                      foundNames []string, responded bool, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = "✅ *Отреагировал* (вероятность 25%%)"
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (вероятность 25%%)"
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер VitalikViebat*\n\n" +
        "%s\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные имена:* %v\n" +
        "📊 *Всего имён:* %d\n" +
        "💬 *Ответ:* %s",
        reactionStatus,
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        foundNames,
        len(foundNames),
        "Осторожно, может въебать!",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога VitalikViebat: %v", err)
    }
}
