package mybot

import (
    "database/sql"
    "log"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    _ "github.com/lib/pq"
)

// SaveMessageToDB - сохраняет сообщение в БД
func SaveMessageToDB(db *sql.DB, botUsername string, msg *tgbotapi.Message) error {
    if db == nil {
        return nil // БД не настроена - пропускаем
    }
    
    // Используем схему main.messages_log
    query := `
        INSERT INTO main.messages_log (
            created_at, bot_id, user_id, message_id, chat_id,
            bot_username, message_text, user_name, user_username,
            has_sticker, has_photo, has_document, chat_title
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `
    
    // Определяем наличие медиа
    hasSticker := msg.Sticker != nil
    hasPhoto := msg.Photo != nil && len(msg.Photo) > 0
    hasDocument := msg.Document != nil
    
    // Получаем chat_title (может быть пустым в личных сообщениях)
    chatTitle := ""
    if msg.Chat.Title != "" {
        chatTitle = msg.Chat.Title
    } else if msg.Chat.UserName != "" {
        // Для личных сообщений используем username
        chatTitle = "@" + msg.Chat.UserName
    } else {
        // Или first_name для приватных чатов
        chatTitle = msg.Chat.FirstName
    }
    
    _, err := db.Exec(query,
        time.Now(),                 // created_at
        0,                          // bot_id (пока 0)
        msg.From.ID,                // user_id
        msg.MessageID,              // message_id  
        msg.Chat.ID,                // chat_id
        botUsername,                // bot_username
        msg.Text,                   // message_text
        msg.From.FirstName,         // user_name
        msg.From.UserName,          // user_username
        hasSticker,                 // has_sticker
        hasPhoto,                   // has_photo
        hasDocument,                // has_document
        chatTitle,                  // chat_title - НОВОЕ ПОЛЕ!
    )
    
    if err != nil {
        log.Printf("❌ Ошибка сохранения в БД: %v", err)
        return err
    }
    
    log.Printf("💾 Сообщение сохранено в БД (чат: %s)", chatTitle)
    return nil
}
