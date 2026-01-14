package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Списки триггерных слов бар/кабак
var barNouns = []string{
    "бар", "бары", "барам", "барах", "баров",
    "кабак", "кабаки", "кабаке", "кабаков",
}

var barPhrases = []string{
    "в баре", "в кабаках",
}

// CheckBarSwinobarmenTriggers проверяет сообщение на слова бар/кабак
// Приоритет: 13-й
// Вероятность: 50% (каждое 2-е примерно)
func CheckBarSwinobarmenTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    text := normalizeText(msg.Text)
    foundWords := []string{}
    
    for _, word := range barNouns {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    for _, phrase := range barPhrases {
        if strings.Contains(text, phrase) {
            foundWords = append(foundWords, phrase)
        }
    }
    
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("🍻 Триггер BarSwinobarmen: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    rand.Seed(time.Now().UnixNano())
    if rand.Float64() > 0.5 {
        log.Printf("🎲 Пропущено BarSwinobarmen (вероятность 50%%)")
        sendBarSwinobarmenTriggerLogToChat(bot, msg, foundWords, false, logChatID)
        return false
    }
    
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "А свинобармен там?")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки BarSwinobarmen: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ BarSwinobarmen")
    sendBarSwinobarmenTriggerLogToChat(bot, msg, foundWords, true, logChatID)
    return true
}

func sendBarSwinobarmenTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                       foundWords []string, responded bool, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = "✅ *Отреагировал* (вероятность 50%%)"
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (вероятность 50%%)"
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер BarSwinobarmen*\n\n" +
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
        "А свинобармен там?",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога BarSwinobarmen: %v", err)
    }
}
