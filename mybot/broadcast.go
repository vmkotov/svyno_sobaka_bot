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
    
    // 🟢 2. ВКЛЮЧЕНИЕ БД - запрос данных с подсчётом
    log.Println("📋 Запрашиваем данные...")
    
    // Сначала посчитаем сколько записей за сегодня
    var totalRecords int
    countQuery := `SELECT COUNT(*) FROM svyno_sobaka_bot.svyno_sobaka_of_the_day WHERE dt_date_only = CURRENT_DATE`
    err = db.QueryRow(countQuery).Scan(&totalRecords)
    if err != nil {
        log.Printf("⚠️ Не удалось подсчитать записи: %v", err)
        totalRecords = 0
    }
    
    log.Printf("📊 В таблице svyno_sobaka_of_the_day найдено %d записей за сегодня", totalRecords)
    
    // Если нет записей - завершаем
    if totalRecords == 0 {
        log.Println("ℹ️ Нет записей для рассылки, завершаю работу")
        return nil
    }
    
    // Запрашиваем детальные данные
    rows, err := db.Query(`
        SELECT 
            chat_id,
            COALESCE(user_username, user_name, 'Аноним') as display_name,
            user_name,
            user_username
        FROM svyno_sobaka_bot.svyno_sobaka_of_the_day 
        WHERE dt_date_only = CURRENT_DATE
        ORDER BY chat_id
    `)
    
    if err != nil {
        log.Printf("❌ Ошибка запроса данных: %v", err)
        return err
    }
    
    // 🔴 3. ВЫКЛЮЧЕНИЕ БД - сразу после получения данных
    defer rows.Close()
    log.Println("✅ Данные получены, БД можно закрывать")
    
    // Теперь работаем только с данными в памяти
    sentCount := 0
    failedCount := 0
    chatIDs := make([]int64, 0)
    
    // Сначала соберём все chat_id для логирования
    tempRows, err := db.Query(`
        SELECT chat_id 
        FROM svyno_sobaka_bot.svyno_sobaka_of_the_day 
        WHERE dt_date_only = CURRENT_DATE
        ORDER BY chat_id
    `)
    if err == nil {
        defer tempRows.Close()
        for tempRows.Next() {
            var chatID int64
            if err := tempRows.Scan(&chatID); err == nil {
                chatIDs = append(chatIDs, chatID)
                log.Printf("📍 Начинаю рассылку в чат %d", chatID)
            }
        }
    }
    
    log.Printf("📍 Всего чатов для рассылки: %d", len(chatIDs))
    
    // Теперь обрабатываем основную выборку
    for rows.Next() {
        var chatID int64
        var displayName, userName, userUsername sql.NullString
        
        if err := rows.Scan(&chatID, &displayName, &userName, &userUsername); err != nil {
            log.Printf("⚠️ Ошибка чтения данных для чата: %v", err)
            failedCount++
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
        
        log.Printf("💬 Чат %d: выбрана свинособака %s", chatID, finalName)
        
        // 1. Первое сообщение
        msg1 := tgbotapi.NewMessage(chatID, "🔍 *Идёт сканирование пользователей чата на наличие свинособаки*")
        msg1.ParseMode = "Markdown"
        
        if _, err := bot.Send(msg1); err != nil {
            log.Printf("⚠️ Не отправилось 1-е сообщение в %d: %v", chatID, err)
            failedCount++
            continue
        }
        
        // Пауза для эффекта
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
            failedCount++
            continue
        }
        
        sentCount++
        log.Printf("✅ Успешно отправлено в чат %d", chatID)
        
        // Пауза между чатами
        time.Sleep(500 * time.Millisecond)
    }
    
    // 🔴 4. ВЫКЛЮЧЕНИЕ БД - проверка ошибок
    if err := rows.Err(); err != nil {
        log.Printf("⚠️ Ошибка при итерации rows: %v", err)
    }
    
    log.Printf("🎉 Рассылка завершена. Статистика:")
    log.Printf("   Всего записей в таблице: %d", totalRecords)
    log.Printf("   Чатов для рассылки: %d", len(chatIDs))
    log.Printf("   Успешно отправлено: %d", sentCount)
    log.Printf("   Не удалось отправить: %d", failedCount)
    
    // Проверяем несоответствие
    if sentCount+failedCount != len(chatIDs) {
        log.Printf("⚠️ Внимание: несоответствие в количестве! sent(%d) + failed(%d) != chats(%d)", 
            sentCount, failedCount, len(chatIDs))
    }
    
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
