package mybot

import (
    "fmt"
    "log"
    "strings"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CheckHuiNadenemTriggers проверяет сообщение на наличие фразы "спартак куда денем"
// (поиск как части строки, без учёта регистра) и отвечает "НА ХУЙ НАДЕНЕМ"
func CheckHuiNadenemTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    textLower := strings.ToLower(msg.Text)
    triggerPhrase := "спартак куда денем"
    
    if !strings.Contains(textLower, triggerPhrase) {
        return false
    }
    
    log.Printf("🔍 Найден триггер 'Спартак куда денем' в сообщении от @%s", 
               msg.From.UserName)
    
    response := "НА ХУЙ НАДЕНЕМ"
    
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки реплая 'Спартак куда денем': %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ на триггер 'Спартак куда денем': %s", response)
    
    sendHuiNadenemTriggerLogToChat(bot, msg, logChatID)
    
    return true
}

func sendHuiNadenemTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) {
    logText := fmt.Sprintf(
        "🔔 *Триггер: Спартак куда денем*\n\n" +
        "✅ *Отреагировал*\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденная фраза:* `%s`\n" +
        "💬 *Ответ:* %s",
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        "спартак куда денем",
        "НА ХУЙ НАДЕНЕМ",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога триггера 'Спартак куда денем': %v", err)
    }
}
