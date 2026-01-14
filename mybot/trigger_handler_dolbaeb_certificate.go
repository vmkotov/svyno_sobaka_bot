package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Список триггерных глаголов (в нижнем регистре)
var dolbaebVerbs = []string{
    "упал",
    "ёбнулся", 
    "пизданулся",
    "ударился",
    "въебался",
    "промахнулся",
}

// CheckDolbaebCertificateTriggers проверяет сообщение на глаголы падения/удара
// Приоритет: 8-й (самый последний)
// Вероятность: 50% (каждое 2-е примерно)
func CheckDolbaebCertificateTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Нормализуем текст
    text := normalizeText(msg.Text)
    
    // Проверяем триггерные глаголы
    foundVerbs := []string{}
    for _, verb := range dolbaebVerbs {
        if strings.Contains(text, verb) {
            foundVerbs = append(foundVerbs, verb)
        }
    }
    
    // Если ничего не найдено
    if len(foundVerbs) == 0 {
        return false
    }
    
    log.Printf("🤕 Триггер DolbaebCertificate: найдено %d глаголов от @%s", 
               len(foundVerbs), msg.From.UserName)
    
    // Проверяем вероятность (50%)
    rand.Seed(time.Now().UnixNano())
    if rand.Float64() > 0.5 { // 50% шанс пропустить
        log.Printf("🎲 Пропущено DolbaebCertificate (вероятность 50%%)")
        sendDolbaebCertificateTriggerLogToChat(bot, msg, foundVerbs, false, logChatID)
        return false
    }
    
    // Отправляем ответ
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "Сертификат долбаёба ему!")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки DolbaebCertificate: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ DolbaebCertificate")
    
    // Логируем
    sendDolbaebCertificateTriggerLogToChat(bot, msg, foundVerbs, true, logChatID)
    
    return true
}

// sendDolbaebCertificateTriggerLogToChat логирует срабатывание триггера
func sendDolbaebCertificateTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                           foundVerbs []string, responded bool, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = "✅ *Отреагировал* (вероятность 50%%)"
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (вероятность 50%%)"
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер DolbaebCertificate*\n\n" +
        "%s\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные глаголы:* %v\n" +
        "📊 *Всего глаголов:* %d\n" +
        "💬 *Ответ:* %s",
        reactionStatus,
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        foundVerbs,
        len(foundVerbs),
        "Сертификат долбаёба ему!",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога DolbaebCertificate: %v", err)
    }
}
