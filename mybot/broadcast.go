package mybot

import (
    "database/sql"
    "log"
    "math/rand"
    "net/http"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SetupBroadcastHandler создаёт HTTP обработчик для рассылки
func SetupBroadcastHandler(bot *tgbotapi.BotAPI, db *sql.DB, secretKey string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Проверяем авторизацию
        if !isAuthorized(r, secretKey) {
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

// isAuthorized проверяет авторизацию
func isAuthorized(r *http.Request, secretKey string) bool {
    // 1. Разрешаем локальные запросы (от Yandex Cloud внутри контейнера)
    if strings.HasPrefix(r.RemoteAddr, "127.0.0.1") || 
       strings.HasPrefix(r.RemoteAddr, "[::1]") {
        return true
    }
    
    // 2. Разрешаем по секретному заголовку (для ручных вызовов)
    if r.Header.Get("X-Broadcast-Secret") == secretKey {
        return true
    }
    
    // 3. Разрешаем по User-Agent Yandex Cloud (если прямой вызов)
    userAgent := strings.ToLower(r.UserAgent())
    if strings.Contains(userAgent, "yandex") || 
       strings.Contains(userAgent, "cloud") {
        return true
    }
    
    return false
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
        // ============ СООБЩЕНИЕ 1 ============
        msg1 := tgbotapi.NewMessage(chatID, "Идёт сканирование пользователей чата на наличие свинособаки...")
        msg1.ParseMode = "Markdown"
        
        if _, err := bot.Send(msg1); err != nil {
            log.Printf("⚠️ Не удалось отправить сообщение 1 в %d: %v", chatID, err)
            continue
        }
        
        // Пауза 3 секунды
        time.Sleep(3 * time.Second)
        
        // ============ ПОЛУЧАЕМ СЛУЧАЙНОГО ПОЛЬЗОВАТЕЛЯ ============
        randomUser := getRandomUserFromChat(db, chatID, botUsername)
        
        // ============ СООБЩЕНИЕ 2 ============
        msg2Text := "«Сегодня свинособака – это " + randomUser + "»"
        msg2 := tgbotapi.NewMessage(chatID, msg2Text)
        msg2.ParseMode = "Markdown"
        
        if _, err := bot.Send(msg2); err != nil {
            log.Printf("⚠️ Не удалось отправить сообщение 2 в %d: %v", chatID, err)
            continue
        }
        
        sentCount++
        
        // Логируем прогресс
        if sentCount%10 == 0 {
            log.Printf("📤 Отправлено %d/%d чатов", sentCount, len(chatIDs))
        }
        
        // Пауза между чатами
        time.Sleep(500 * time.Millisecond)
    }
    
    log.Printf("🎉 Рассылка завершена. Успешно: %d/%d", sentCount, len(chatIDs))
    return nil
}

// getRandomUserFromChat выбирает случайного пользователя из чата
func getRandomUserFromChat(db *sql.DB, chatID int64, botUsername string) string {
    // Запрашиваем всех пользователей из этого чата
    rows, err := db.Query(`
        SELECT DISTINCT user_name, user_username 
        FROM main.messages_log 
        WHERE chat_id = $1 
        AND bot_username = $2
        AND user_name IS NOT NULL
        AND user_name != ''
    `, chatID, botUsername)
    
    if err != nil {
        log.Printf("⚠️ Ошибка запроса пользователей для чата %d: %v", chatID, err)
        return "неизвестный пользователь"
    }
    defer rows.Close()
    
    // Собираем пользователей
    var users []string
    for rows.Next() {
        var name, username sql.NullString
        if err := rows.Scan(&name, &username); err != nil {
            continue
        }
        
        if username.Valid && username.String != "" {
            users = append(users, "@"+username.String)
        } else if name.Valid && name.String != "" {
            users = append(users, name.String)
        }
    }
    
    // Если пользователей нет
    if len(users) == 0 {
        return "неизвестный пользователь"
    }
    
    // Выбираем случайного
    rand.Seed(time.Now().UnixNano())
    randomIndex := rand.Intn(len(users))
    
    log.Printf("🎲 Для чата %d выбран пользователь: %s (всего %d пользователей)", 
               chatID, users[randomIndex], len(users))
    
    return users[randomIndex]
}
