package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Списки слов мразь/дебил
var mrazNouns = []string{
    "мразь", "мрази", "мразей", "мразями", "мразота", "мразоты", 
    "мразотами", "мразотство", "тварь", "твари",
    "уебок", "уёбок", "уебки", "уёбки", "уебаны", "уебан",
    "дебил", "дебилы", "дебилам", "дебилов",
    "идиот", "идиоты", "идиотам", "идиотов",
    "еблан", "ёбобо", "ебобо", "ебанько",
}

var mrazAdjectives = []string{
    "тупой", "тупая", "тупые",
    "конченый", "конченая", "конченые",
    "тупорылый", "тупорылая", "тупорылые",
}

// Варианты ответов
var mrazResponses = []string{
    "Мразотство! Всё как мы любим!",
    "Лучше быть просто свинособакой",
}

// CheckMrazotstvoTriggers проверяет сообщение на мразь/дебил
// Приоритет: 16-й
// Вероятность: 25% + случайный выбор из 2 вариантов
func CheckMrazotstvoTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    text := normalizeText(msg.Text)
    foundWords := []string{}
    
    for _, word := range mrazNouns {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    for _, word := range mrazAdjectives {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("👿 Триггер Mrazotstvo: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    rand.Seed(time.Now().UnixNano())
    if rand.Float64() > 0.25 { // 75% шанс пропустить
        log.Printf("🎲 Пропущено Mrazotstvo (вероятность 25%%)")
        sendMrazotstvoTriggerLogToChat(bot, msg, foundWords, false, 0, logChatID)
        return false
    }
    
    // Выбираем случайный ответ
    responseIndex := rand.Intn(len(mrazResponses))
    response := mrazResponses[responseIndex]
    
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки Mrazotstvo: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ Mrazotstvo: вариант %d", responseIndex+1)
    sendMrazotstvoTriggerLogToChat(bot, msg, foundWords, true, responseIndex, logChatID)
    return true
}

func sendMrazotstvoTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                   foundWords []string, responded bool, responseIndex int, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = fmt.Sprintf("✅ *Отреагировал* (вероятность 25%%, вариант %d/%d)", 
                                    responseIndex+1, len(mrazResponses))
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (вероятность 25%%)"
    }
    
    responseText := ""
    if responded {
        responseText = mrazResponses[responseIndex]
    } else {
        responseText = mrazResponses[0] + " или " + mrazResponses[1]
    }
    
    // Обрезаем если много слов
    wordsForLog := foundWords
    if len(foundWords) > 5 {
        wordsForLog = foundWords[:5]
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер Mrazotstvo*\n\n" +
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
        wordsForLog,
        len(foundWords),
        responseText,
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога Mrazotstvo: %v", err)
    }
}
