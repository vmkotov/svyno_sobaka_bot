package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Список триггерных слов (в нижнем регистре)
var svinoSobakaV2Words = []string{
    "свинособака",
    "свинособаки", 
    "свинособакам",
    "свинособак",
    "свинособачник",
}

// CheckSvinoSobakaV2Triggers проверяет сообщение на слова свинособака-v2
// Приоритет: 7-й
// Вероятность: 33% (примерно каждое 3-е)
func CheckSvinoSobakaV2Triggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Нормализуем текст
    text := normalizeText(msg.Text)
    
    // Проверяем триггерные слова
    foundWords := []string{}
    for _, word := range svinoSobakaV2Words {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Если ничего не найдено
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("🐷 Триггер SvinoSobakaV2: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    // Проверяем вероятность (33%)
    rand.Seed(time.Now().UnixNano())
    if rand.Float64() > 0.33 { // 67% шанс пропустить
        log.Printf("🎲 Пропущено SvinoSobakaV2 (вероятность 33%%)")
        sendSvinoSobakaV2TriggerLogToChat(bot, msg, foundWords, false, logChatID)
        return false
    }
    
    // Отправляем ответ
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "А может быть всё-таки свинособака – это ты?")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки SvinoSobakaV2: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ SvinoSobakaV2")
    
    // Логируем
    sendSvinoSobakaV2TriggerLogToChat(bot, msg, foundWords, true, logChatID)
    
    return true
}

// sendSvinoSobakaV2TriggerLogToChat логирует срабатывание триггера
func sendSvinoSobakaV2TriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                      foundWords []string, responded bool, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = "✅ *Отреагировал* (вероятность 33%%)"
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (вероятность 33%%)"
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер SvinoSobakaV2*\n\n" +
        "%s\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные слова:* %v\n" +
        "📊 *Всего слов:* %d\n" +
        "💬 *Ответ:* %s",
        reactionStatus,
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        foundWords,
        len(foundWords),
        "А может быть всё-таки свинособака – это ты?",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога SvinoSobakaV2: %v", err)
    }
}
