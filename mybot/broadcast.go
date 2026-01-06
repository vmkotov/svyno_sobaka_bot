package mybot

import (
    "database/sql"
    "log"
    "net/http"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SetupBroadcastHandler создаёт HTTP обработчик для рассылки
func SetupBroadcastHandler(bot *tgbotapi.BotAPI, db *sql.DB, secretKey string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Проверяем авторизацию (по User-Agent от Yandex Cloud)
        if !isAuthorized(r) {
            log.Printf("❌ Неавторизованный запрос от %s, User-Agent: %s", 
                      r.RemoteAddr, r.UserAgent())
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        log.Printf("🔔 Запуск рассылки по запросу от %s", r.RemoteAddr)
        
        // 2. Запускаем рассылку в фоне
        go func() {
            if err := SendBroadcast(bot, db, "svyno_sobaka_bot"); err != nil {
                log.Printf("❌ Ошибка рассылки: %v", err)
            }
        }()
        
        // 3. Отвечаем сразу
        w.WriteHeader(http.StatusAccepted)
        w.Write([]byte("Broadcast started"))
    }
}

// isAuthorized проверяет что запрос от Yandex Cloud
func isAuthorized(r *http.Request) bool {
    // Разрешаем запросы с User-Agent содержащим "Yandex" или "cloud"
    userAgent := strings.ToLower(r.UserAgent())
    return strings.Contains(userAgent, "yandex") || 
           strings.Contains(userAgent, "cloud") ||
           strings.Contains(r.RemoteAddr, "10.") || // Внутренние IP Yandex Cloud
           r.Header.Get("X-Broadcast-Secret") == "change-me-in-production" // Ручные запросы
}

// SendBroadcast выполняет рассылку по всем чатам из БД
func SendBroadcast(bot *tgbotapi.BotAPI, db *sql.DB, botUsername string) error {
    if db == nil {
        log.Println("ℹ️ БД не настроена, пропускаем рассылку")
        return nil
    }
    
    log.Println("📢 Начинаю рассылку...")
    
    // 1. Берём уникальные chat_id
    rows, err := db.Query(`
        SELECT DISTINCT chat_id 
        FROM main.messages_log 
        WHERE chat_id IS NOT NULL 
        AND bot_username = $1
        AND chat_id != 0
        ORDER BY chat_id
    `, botUsername)
    
    if err != nil {
        log.Printf("❌ Ошибка запроса к БД: %v", err)
        return err
    }
    defer rows.Close()
    
    // 2. Собираем chat_id
    var chatIDs []int64
    for rows.Next() {
        var chatID int64
        if err := rows.Scan(&chatID); err != nil {
            log.Printf("⚠️ Ошибка чтения chat_id: %v", err)
            continue
        }
        chatIDs = append(chatIDs, chatID)
    }
    
    if len(chatIDs) == 0 {
        log.Println("ℹ️ Нет chat_id для рассылки")
        return nil
    }
    
    log.Printf("📋 Найдено %d чатов для рассылки", len(chatIDs))
    
    // 3. Отправляем каждому чату
    sentCount := 0
    for _, chatID := range chatIDs {
        msg := tgbotapi.NewMessage(chatID,
            "📢 *ЕЖЕДНЕВНОЕ СООБЩЕНИЕ ОТ СВИНОСОБАКИ*\n\n" +
            "Не забывай писать мне сообщения!\n" +
            "Используй /start для начала\n" +
            "И /help для помощи\n\n" +
            "А пока иди нахуй! 🐷🐶")
        
        msg.ParseMode = "Markdown"
        
        if _, err := bot.Send(msg); err != nil {
            log.Printf("⚠️ Не удалось отправить в %d: %v", chatID, err)
            continue
        }
        
        sentCount++
        
        // Пауза между сообщениями
        if sentCount%10 == 0 {
            log.Printf("📤 Отправлено %d/%d", sentCount, len(chatIDs))
        }
        
        time.Sleep(100 * time.Millisecond)
    }
    
    log.Printf("🎉 Рассылка завершена. Успешно: %d/%d", sentCount, len(chatIDs))
    return nil
}
