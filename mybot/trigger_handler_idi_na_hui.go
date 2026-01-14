package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Список триггерных фраз (в нижнем регистре)
var idiNaHuiPhrases = []string{
    "иди на хуй", "идите на хуй", "идут на хуй", "идет на хуй", "идёт на хуй",
    "на хуй идет", "на хуй идёт", "на хуй", "нахуй",
    "пошла на хуй", "пошла она на хуй", "идите все на хуй",
    "пошла в пизду", "пошел в пизду", "пошёл в пизду",
    "пошла она в пизду", "пошли в пизду", "пошли они в пизду",
    "пошли они все в пизду", "в пизду",
    "послал на хуй", "пошлет на хуй", "пошлёт на хуй",
    "послан на хуй", "посланы на хуй",
}

// CheckIdiNaHuiTriggers проверяет сообщение на фразы "иди на хуй"
// Приоритет: 12-й
// Вероятность: 33% (каждое 3-е примерно)
func CheckIdiNaHuiTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Нормализуем текст
    text := normalizeText(msg.Text)
    
    // Проверяем триггерные фразы
    foundPhrases := []string{}
    for _, phrase := range idiNaHuiPhrases {
        if strings.Contains(text, phrase) {
            foundPhrases = append(foundPhrases, phrase)
        }
    }
    
    // Если ничего не найдено
    if len(foundPhrases) == 0 {
        return false
    }
    
    log.Printf("👊 Триггер IdiNaHui: найдено %d фраз от @%s", 
               len(foundPhrases), msg.From.UserName)
    
    // Проверяем вероятность (33%)
    rand.Seed(time.Now().UnixNano())
    if rand.Float64() > 0.33 { // 67% шанс пропустить
        log.Printf("🎲 Пропущено IdiNaHui (вероятность 33%%)")
        sendIdiNaHuiTriggerLogToChat(bot, msg, foundPhrases, false, logChatID)
        return false
    }
    
    // Отправляем ответ
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "Иди-ка лучше ты на хуй!")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки IdiNaHui: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ IdiNaHui")
    
    // Логируем
    sendIdiNaHuiTriggerLogToChat(bot, msg, foundPhrases, true, logChatID)
    
    return true
}

// sendIdiNaHuiTriggerLogToChat логирует срабатывание триггера
func sendIdiNaHuiTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                 foundPhrases []string, responded bool, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = "✅ *Отреагировал* (вероятность 33%%)"
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (вероятность 33%%)"
    }
    
    // Обрезаем список если их много
    phrasesForLog := foundPhrases
    if len(foundPhrases) > 3 {
        phrasesForLog = foundPhrases[:3]
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер IdiNaHui*\n\n" +
        "%s\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные фразы:* %v\n" +
        "📊 *Всего фраз:* %d\n" +
        "💬 *Ответ:* %s",
        reactionStatus,
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        phrasesForLog,
        len(foundPhrases),
        "Иди-ка лучше ты на хуй!",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога IdiNaHui: %v", err)
    }
}
