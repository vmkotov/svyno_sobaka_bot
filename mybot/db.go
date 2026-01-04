package mybot

import (
	"database/sql"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

// SaveMessageToDB - просто сохраняет сообщение в БД
func SaveMessageToDB(db *sql.DB, botUsername string, msg *tgbotapi.Message) error {
	if db == nil {
		return nil // БД не настроена - пропускаем
	}

	query := `
        INSERT INTO messages_log (
            created_at, bot_id, user_id, message_id, chat_id,
            bot_username, message_text, user_name, user_username
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `

	_, err := db.Exec(query,
		time.Now(),         // created_at
		0,                  // bot_id (пока 0)
		msg.From.ID,        // user_id
		msg.MessageID,      // message_id
		msg.Chat.ID,        // chat_id
		botUsername,        // bot_username
		msg.Text,           // message_text
		msg.From.FirstName, // user_name
		msg.From.UserName,  // user_username
	)

	if err != nil {
		log.Printf("❌ Ошибка сохранения в БД: %v", err)
		return err
	}

	log.Printf("💾 Сообщение сохранено в БД")
	return nil
}
