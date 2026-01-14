package mybot

import (
    "fmt"
    "log"
    "strings"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Списки триггерных слов Россия/родина (в нижнем регистре)
var russiaNouns = []string{
    "россия", "россии", "россией", "россию",
    "россияне", "россиянам",
    "рф",
    "родина", "родине", "родиной", "родину", "родины",
}

var russiaAdjectives = []string{
    "русский", "русским",
    "российский", "российских", "российские",
    "обрусевший",
}

var russiaPhrases = []string{
    "в нашей стране",
    "у нас в стране", 
    "на нашей с вами родине",
}

// CheckRussiaSlavaTriggers проверяет сообщение на слова Россия/родина
// Приоритет: 10-й (самый последний)
// Вероятность: 100% (всегда)
func CheckRussiaSlavaTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Нормализуем текст
    text := normalizeText(msg.Text)
    
    // Проверяем все триггерные слова и фразы
    foundWords := []string{}
    
    // Проверяем существительные
    for _, word := range russiaNouns {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Проверяем прилагательные
    for _, word := range russiaAdjectives {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Проверяем словосочетания
    for _, phrase := range russiaPhrases {
        if strings.Contains(text, phrase) {
            foundWords = append(foundWords, phrase)
        }
    }
    
    // Если ничего не найдено
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("🇷🇺 Триггер RussiaSlava: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    // Отправляем ответ (всегда)
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "Слава России!")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки RussiaSlava: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ RussiaSlava")
    
    // Логируем
    sendRussiaSlavaTriggerLogToChat(bot, msg, foundWords, logChatID)
    
    return true
}

// sendRussiaSlavaTriggerLogToChat логирует срабатывание триггера
func sendRussiaSlavaTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                    foundWords []string, logChatID int64) {
    
    // Обрезаем список слов если их много
    wordsForLog := foundWords
    if len(foundWords) > 5 {
        wordsForLog = foundWords[:5]
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер RussiaSlava*\n\n" +
        "✅ *Отреагировал* (всегда 100%%)\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные слова:* %v\n" +
        "📊 *Всего слов:* %d\n" +
        "💬 *Ответ:* %s",
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        wordsForLog,
        len(foundWords),
        "Слава России!",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога RussiaSlava: %v", err)
    }
}
