package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Списки слов вчера/помните
var vcheraWords = []string{
    "вчера", "позавчера", "сегодня", "завтра",
}

var vcheraPhrases = []string{
    "как в тот раз", "а помните", "вспомни", "вспомню", 
    "вспомнишь", "не забудь", "не забудьте", "не забываем",
}

// Варианты ответов
var vcheraResponses = []string{
    "Это да, а кто вчера опять нажрался?",
    "Это да, а вчера кто нажрался как свинотавр?",
}

// CheckVcheraNajralsyaTriggers проверяет сообщение на вчера/помните
// Приоритет: 15-й
// Вероятность: 33% + случайный выбор из 2 вариантов
func CheckVcheraNajralsyaTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    text := normalizeText(msg.Text)
    foundWords := []string{}
    
    for _, word := range vcheraWords {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    for _, phrase := range vcheraPhrases {
        if strings.Contains(text, phrase) {
            foundWords = append(foundWords, phrase)
        }
    }
    
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("🕰️ Триггер VcheraNajralsya: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    rand.Seed(time.Now().UnixNano())
    if rand.Float64() > 0.33 { // 67% шанс пропустить
        log.Printf("🎲 Пропущено VcheraNajralsya (вероятность 33%%)")
        sendVcheraNajralsyaTriggerLogToChat(bot, msg, foundWords, false, 0, logChatID)
        return false
    }
    
    // Выбираем случайный ответ
    responseIndex := rand.Intn(len(vcheraResponses))
    response := vcheraResponses[responseIndex]
    
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки VcheraNajralsya: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ VcheraNajralsya: вариант %d", responseIndex+1)
    sendVcheraNajralsyaTriggerLogToChat(bot, msg, foundWords, true, responseIndex, logChatID)
    return true
}

func sendVcheraNajralsyaTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                        foundWords []string, responded bool, responseIndex int, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = fmt.Sprintf("✅ *Отреагировал* (вероятность 33%%, вариант %d/%d)", 
                                    responseIndex+1, len(vcheraResponses))
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (вероятность 33%%)"
    }
    
    responseText := ""
    if responded {
        responseText = vcheraResponses[responseIndex]
    } else {
        responseText = vcheraResponses[0] + " или " + vcheraResponses[1]
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер VcheraNajralsya*\n\n" +
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
        responseText,
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога VcheraNajralsya: %v", err)
    }
}
