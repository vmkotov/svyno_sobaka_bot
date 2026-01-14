package mybot

import (
    "fmt"
    "log"
    "strings"
    "sync"
    
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

// Глобальный счетчик для статистики по чатам
var svinoSobakaV2Counters = make(map[int64]int) // chatID -> counter
var svinoSobakaV2Mutex sync.Mutex

// CheckSvinoSobakaV2Triggers проверяет сообщение на слова свинособака-v2
// Приоритет: 7-й (самый последний)
// Реагирует на каждое 3-е срабатывание в чате
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
    
    log.Printf("🐷 Триггер SvinoSobakaV2: найдено %d слов от @%s в чате %d", 
               len(foundWords), msg.From.UserName, msg.Chat.ID)
    
    // Блокируем для безопасного доступа к счетчику
    svinoSobakaV2Mutex.Lock()
    
    // Инициализируем счетчик для чата если нужно
    if _, exists := svinoSobakaV2Counters[msg.Chat.ID]; !exists {
        svinoSobakaV2Counters[msg.Chat.ID] = 0
    }
    
    // Увеличиваем счетчик
    svinoSobakaV2Counters[msg.Chat.ID]++
    counter := svinoSobakaV2Counters[msg.Chat.ID]
    
    svinoSobakaV2Mutex.Unlock()
    
    // Определяем нужно ли отправлять ответ (каждое 3-е)
    shouldRespond := (counter % 3 == 0)
    
    if shouldRespond {
        // Отправляем ответ
        replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "А может быть всё-таки свинособака – это ты?")
        replyMsg.ReplyToMessageID = msg.MessageID
        
        if _, err := bot.Send(replyMsg); err != nil {
            log.Printf("❌ Ошибка отправки SvinoSobakaV2: %v", err)
            return false
        }
        
        log.Printf("✅ Отправлен ответ SvinoSobakaV2 (счётчик: %d)", counter)
    } else {
        log.Printf("🎲 Пропущено SvinoSobakaV2 (счётчик: %d, ждём 3)", counter)
    }
    
    // Логируем (всегда, даже если не отправили ответ)
    sendSvinoSobakaV2TriggerLogToChat(bot, msg, foundWords, counter, shouldRespond, logChatID)
    
    return shouldRespond // Возвращаем true только если отправили ответ
}

// sendSvinoSobakaV2TriggerLogToChat логирует срабатывание триггера
func sendSvinoSobakaV2TriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                      foundWords []string, counter int, responded bool, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = "✅ *Отреагировал* (каждое 3-е)"
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (счётчик не кратен 3)"
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер SvinoSobakaV2*\n\n" +
        "%s\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные слова:* %v\n" +
        "📊 *Всего слов:* %d\n" +
        "🔢 *Счётчик в чате:* %d\n" +
        "🎯 *Нужно для реакции:* каждое 3-е\n" +
        "💬 *Ответ:* %s",
        reactionStatus,
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        foundWords,
        len(foundWords),
        counter,
        "А может быть всё-таки свинособака – это ты?",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога SvinoSobakaV2: %v", err)
    }
}
