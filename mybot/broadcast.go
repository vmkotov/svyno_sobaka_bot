package mybot

import (
    "database/sql"
    "log"
    "net/http"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SetupBroadcastHandler создаёт HTTP обработчик для рассылки
func SetupBroadcastHandler(bot *tgbotapi.BotAPI, db *sql.DB, secretKey string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Проверяем авторизацию
        if !isAuthorized(r, secretKey) {
            log.Println("❌ Неавторизованный запрос на рассылку")
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        log.Println("🔔 Запуск рассылки по запросу от", r.RemoteAddr)
        
        // 2. Запускаем рассылку в фоне (не блокируем ответ)
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

// isAuthorized проверяет секретный ключ
func isAuthorized(r *http.Request, secretKey string) bool {
    receivedKey := r.Header.Get("X-Broadcast-Secret")
    return receivedKey == secretKey
}

// SendBroadcast выполняет рассылку по всем чатам из БД
func SendBroadcast(bot *tgbotapi.BotAPI, db *sql.DB, botUsername string) error {
    if db == nil {
        return nil // БД не настроена
    }
    
    log.Println("📢 Начинаю ежедневную рассылку...")
    
    // 1. Берём ВСЕ уникальные chat_id где bot_username = 'svyno_sobaka_bot'
    rows, err := db.Query(`
        SELECT DISTINCT chat_id 
        FROM main.messages_log 
        WHERE chat_id IS NOT NULL 
        AND bot_username = $1
        ORDER BY chat_id
    `, botUsername)
    
    if err != nil {
        log.Printf("❌ Ошибка запроса к БД: %v", err)
        return err
    }
    defer rows.Close()
    
    // 2. Собираем все chat_id
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
        // Пропускаем отрицательные ID (группы/каналы) если нужно
        // if chatID < 0 { continue } 
        
        // Формируем сообщение
        msg := tgbotapi.NewMessage(chatID,
            "📢 *ЕЖЕДНЕВНОЕ СООБЩЕНИЕ ОТ СВИНОСОБАКИ*\n\n" +
            "Не забывай писать мне сообщения!\n" +
            "Используй /start для начала\n" +
            "И /help для помощи\n\n" +
            "А пока иди нахуй! 🐷🐶")
        
        msg.ParseMode = "Markdown"
        
        // Отправляем с обработкой ошибок
        if _, err := bot.Send(msg); err != nil {
            log.Printf("⚠️ Не удалось отправить в %d: %v", chatID, err)
            // Можно пропустить или остановиться
            continue
        }
        
        sentCount++
        log.Printf("✅ Отправлено в чат %d (%d/%d)", chatID, sentCount, len(chatIDs))
        
        // Пауза между сообщениями (лимиты Telegram: ~30 сообщений/секунду)
        time.Sleep(100 * time.Millisecond) // 10 сообщений/секунду - безопасно
    }
    
    // 4. Сохраняем лог рассылки в БД
    saveBroadcastLog(db, botUsername, sentCount, len(chatIDs))
    
    log.Printf("🎉 Рассылка завершена. Успешно: %d/%d", sentCount, len(chatIDs))
    return nil
}

// saveBroadcastLog сохраняет информацию о рассылке
func saveBroadcastLog(db *sql.DB, botUsername string, sent, total int) {
    if db == nil {
        return
    }
    
    query := `
        INSERT INTO main.broadcast_log 
        (bot_username, sent_count, total_count, created_at) 
        VALUES ($1, $2, $3, $4)
    `
    
    _, err := db.Exec(query, botUsername, sent, total, time.Now())
    if err != nil {
        log.Printf("⚠️ Не удалось сохранить лог рассылки: %v", err)
        // Можно создать таблицу если нет
    }
}
