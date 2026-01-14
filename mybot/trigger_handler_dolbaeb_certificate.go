package mybot

import (
    "fmt"
    "log"
    "strings"
    "sync"
    
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

// Глобальный счетчик для статистики по чатам
var dolbaebCounters = make(map[int64]int) // chatID -> counter
var dolbaebMutex sync.Mutex

// CheckDolbaebCertificateTriggers проверяет сообщение на глаголы падения/удара
// Приоритет: 8-й (самый последний)
// Реагирует на каждое 2-е срабатывание в чате
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
    
    log.Printf("🤕 Триггер DolbaebCertificate: найдено %d глаголов от @%s в чате %d", 
               len(foundVerbs), msg.From.UserName, msg.Chat.ID)
    
    // Блокируем для безопасного доступа к счетчику
    dolbaebMutex.Lock()
    
    // Инициализируем счетчик для чата если нужно
    if _, exists := dolbaebCounters[msg.Chat.ID]; !exists {
        dolbaebCounters[msg.Chat.ID] = 0
    }
    
    // Увеличиваем счетчик
    dolbaebCounters[msg.Chat.ID]++
    counter := dolbaebCounters[msg.Chat.ID]
    
    dolbaebMutex.Unlock()
    
    // Определяем нужно ли отправлять ответ (каждое 2-е)
    shouldRespond := (counter % 2 == 0)
    
    if shouldRespond {
        // Отправляем ответ
        replyMsg := tgbotapi.NewMessage(msg.Chat.ID, "Сертификат долбаёба ему!")
        replyMsg.ReplyToMessageID = msg.MessageID
        
        if _, err := bot.Send(replyMsg); err != nil {
            log.Printf("❌ Ошибка отправки DolbaebCertificate: %v", err)
            return false
        }
        
        log.Printf("✅ Отправлен ответ DolbaebCertificate (счётчик: %d)", counter)
    } else {
        log.Printf("🎲 Пропущено DolbaebCertificate (счётчик: %d, ждём 2)", counter)
    }
    
    // Логируем (всегда, даже если не отправили ответ)
    sendDolbaebCertificateTriggerLogToChat(bot, msg, foundVerbs, counter, shouldRespond, logChatID)
    
    return shouldRespond // Возвращаем true только если отправили ответ
}

// sendDolbaebCertificateTriggerLogToChat логирует срабатывание триггера
func sendDolbaebCertificateTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                                           foundVerbs []string, counter int, responded bool, logChatID int64) {
    
    var reactionStatus string
    if responded {
        reactionStatus = "✅ *Отреагировал* (каждое 2-е)"
    } else {
        reactionStatus = "🎲 *Пропущено рандомайзером* (счётчик не кратен 2)"
    }
    
    logText := fmt.Sprintf(
        "🔔 *Триггер DolbaebCertificate*\n\n" +
        "%s\n" +
        "📝 *Сообщение:* `%s`\n" +
        "👤 *Пользователь:* %s\n" +
        "💬 *Чат ID:* `%d`\n" +
        "🎯 *Найденные глаголы:* %v\n" +
        "📊 *Всего глаголов:* %d\n" +
        "🔢 *Счётчик в чате:* %d\n" +
        "🎯 *Нужно для реакции:* каждое 2-е\n" +
        "💬 *Ответ:* %s",
        reactionStatus,
        escapeMarkdown(msg.Text),
        escapeMarkdown(msg.From.FirstName),
        msg.Chat.ID,
        foundVerbs,
        len(foundVerbs),
        counter,
        "Сертификат долбаёба ему!",
    )
    
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога DolbaebCertificate: %v", err)
    }
}
