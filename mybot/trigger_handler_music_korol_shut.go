package mybot

import (
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Списки триггерных слов музыки (в нижнем регистре)
var musicNouns = []string{
    "музыка", "музыку", "музыкой", "музыки", "музыке",
    "музло", "музла",
    "песня", "песен", "песни", "песням", "песнями",
    "реп", "рэп", "репчик", "рэпчик", "репер", "рэпер",
    "попса", "попсу", "попсы",
    "клип", "клипе",
}

var musicAdjectives = []string{
    "музыкальный", "музыкальной", "музыкальная",
}

// CheckMusicKorolShutTriggers проверяет сообщение на музыкальные слова
// Приоритет: 9-й (самый последний)
// Вероятность: 50% (каждое 2-е примерно)
func CheckMusicKorolShutTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64) bool {
    if msg.Text == "" {
        return false
    }
    
    // Нормализуем текст
    text := normalizeText(msg.Text)
    
    // Проверяем все триггерные слова
    foundWords := []string{}
    
    // Проверяем существительные
    for _, word := range musicNouns {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Проверяем прилагательные
    for _, word := range musicAdjectives {
        if strings.Contains(text, word) {
            foundWords = append(foundWords, word)
        }
    }
    
    // Если ничего не найдено
    if len(foundWords) == 0 {
        return false
    }
    
    log.Printf("🎵 Триггер MusicKorolShut: найдено %d слов от @%s", 
               len(foundWords), msg.From.UserName)
    
    // Проверяем вероятность (50%)
    rand.Seed(time.Now().UnixNano())
    if rand.Float64() > 0.5 { // 50% шанс пропустить
        log.Printf("🎲 Пропущено MusicKorolShut (вероятность 50%%)")
        sendMusicKorolShutTriggerLogToChat(bot, msg, foundWords, false, logChatID)
        return false
    }
    
    // Отправляем ответ
    replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "Лучше включи Король и Шут")
    replyMsg.ReplyToMessageID = msg.MessageID
    
    if _, err := bot.Send(replyMsg); err != nil {
        log.Printf("❌ Ошибка отправки MusicKorolShut: %v", err)
        return false
    }
    
    log.Printf("✅ Отправлен ответ MusicKorolShut")
    
    // Логируем
    sendMusicKorolShutTriggerLogToChat(bot, msg, foundWords, true, logChatID)
    
    return true
}

// sendMusicKorolShutTriggerLogToChat логирует срабатывание триггера
func sendMusicKorolShutTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                       foundWords []string, responded bool, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = "✅ *Отреагировал* (вероятность 50%%)"
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (вероятность 50%%)"
    }
    
    // Обрезаем список слов если их много
    wordsForLog := foundWords
    if len(foundWords) > 5 {
        wordsForLog = foundWords[:5]
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер MusicKorolShut*\n\n" +
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
        "Лучше включи Король и Шут",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога MusicKorolShut: %v", err)
    }
}
