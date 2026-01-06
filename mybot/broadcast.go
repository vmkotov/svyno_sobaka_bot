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
        if !isAuthorized(r, secretKey) {
            log.Printf("❌ Неавторизованный запрос от %s", r.RemoteAddr)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        log.Printf("🔔 Запуск рассылки по запросу от %s", r.RemoteAddr)
        
        go func() {
            if err := SendSvynoSobakaBroadcast(bot, db); err != nil {
                log.Printf("❌ Ошибка рассылки: %v", err)
            }
        }()
        
        w.WriteHeader(http.StatusAccepted)
        w.Write([]byte("Svyno sobaka broadcast started"))
    }
}

// SendSvynoSobakaBroadcast выполняет рассылку с выбором свинособаки дня
func SendSvynoSobakaBroadcast(bot *tgbotapi.BotAPI, db *sql.DB) error {
    if db == nil {
        log.Println("ℹ️ БД не настроена, пропускаем рассылку")
        return nil
    }
    
    log.Println("📢 Начинаю рассылку свинособаки дня...")
    
    // 🟢 1. ВКЛЮЧЕНИЕ БД - вызов процедуры
    log.Println("🔄 Вызываем процедуру...")
    _, err := db.Exec(`CALL svyno_sobaka_bot.proc_svyno_sobaka_of_the_day()`)
    if err != nil {
        log.Printf("❌ Ошибка вызова процедуры: %v", err)
    } else {
        log.Println("✅ Процедура выполнена")
    }
    
    // 🟢 2. ВКЛЮЧЕНИЕ БД - запрос данных
    log.Println("📋 Запрашиваем данные...")
    rows, err := db.Query(`
        SELECT 
            chat_id,
            COALESCE(user_username, user_name, 'Аноним') as display_name,
            user_name,
            user_username
        FROM svyno_sobaka_bot.svyno_sobaka_of_the_day 
        WHERE dt_insert::date = CURRENT_DATE
        ORDER BY chat_id
    `)
    
    if err != nil {
        log.Printf("❌ Ошибка запроса: %v", err)
        return err
    }
    
    // 🔴 3. ВЫКЛЮЧЕНИЕ БД - сразу после получения данных
    defer rows.Close()
    log.Println("✅ Данные получены, БД можно закрывать")
    
    // Теперь работаем только с данными в памяти
    
    sentCount := 0
    for rows.Next() {
        var chatID int64
        var displayName, userName, userUsername sql.NullString
        
        if err := rows.Scan(&chatID, &displayName, &userName, &userUsername); err != nil {
            log.Printf("⚠️ Ошибка чтения: %v", err)
            continue
        }
        
        // Формируем имя
        var finalName string
        if userUsername.Valid && userUsername.String != "" {
            finalName = "@" + userUsername.String
        } else if userName.Valid && userName.String != "" {
            finalName = userName.String
        } else {
            finalName = "Аноним"
        }
        
        log.Printf("💬 Чат %d: %s", chatID, finalName)
        
        // 1. Первое сообщение
        msg1 := tgbotapi.NewMessage(chatID, "🔍 *Идёт сканирование пользователей чата на наличие свинособаки*")
        msg1.ParseMode = "Markdown"
        
        if _, err := bot.Send(msg1); err != nil {
            log.Printf("⚠️ Не отправилось 1-е сообщение в %d: %v", chatID, err)
            continue
        }
        
        // Пауза
        time.Sleep(3 * time.Second)
        
        // 2. Второе сообщение
        msg2 := tgbotapi.NewMessage(chatID,
            "🎉 *СВИНОСОБАКА ДНЯ*\n\n"+
                "Сегодня свинособака – это *"+finalName+"*\n\n"+
                "Поздравляем с этим почётным званием! 🐷🐶\n"+
                "Это безусловно успех 🎊")
        msg2.ParseMode = "Markdown"
        
        if _, err := bot.Send(msg2); err != nil {
            log.Printf("⚠️ Не отправилось 2-е сообщение в %d: %v", chatID, err)
            continue
        }
        
        sentCount++
        log.Printf("✅ Отправлено в чат %d", chatID)
        
        // Пауза между чатами
        time.Sleep(500 * time.Millisecond)
    }
    
    // 🔴 4. ВЫКЛЮЧЕНИЕ БД - проверка ошибок
    if err := rows.Err(); err != nil {
        log.Printf("⚠️ Ошибка rows: %v", err)
    }
    
    // 🔴 5. ВЫКЛЮЧЕНИЕ БД - rows закрываются через defer
    
    log.Printf("🎉 Рассылка завершена. Отправлено: %d", sentCount)
    return nil
}

// isAuthorized проверяет авторизацию
func isAuthorized(r *http.Request, secretKey string) bool {
    if strings.HasPrefix(r.RemoteAddr, "127.0.0.1") || 
       strings.HasPrefix(r.RemoteAddr, "[::1]") {
        return true
    }
    
    if r.Header.Get("X-Broadcast-Secret") == secretKey {
        return true
    }
    
    userAgent := strings.ToLower(r.UserAgent())
    if strings.Contains(userAgent, "yandex") || 
       strings.Contains(userAgent, "cloud") {
        return true
    }
    
    return false
}
