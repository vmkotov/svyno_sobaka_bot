package mybot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

func GetTriggersConfigJSON(db *sql.DB) ([]byte, error) {
	if db == nil {
		return nil, fmt.Errorf("📭 Подключение к БД не настроено")
	}

	log.Printf("🗃️ Загружаю конфигурацию триггеров из БД...")

	// ДОБАВЛЯЕМ TIMEOUT
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var jsonData string
	err := db.QueryRowContext(ctx, "SELECT svyno_sobaka_bot.get_triggers_config_json()").Scan(&jsonData)

	if err != nil {
		log.Printf("❌ Не удалось загрузить триггеры из БД: %v", err)
		return nil, fmt.Errorf("🗄️ Ошибка загрузки триггеров из БД: %w", err)
	}

	log.Printf("✅ Конфигурация триггеров успешно загружена из БД (%d байт)", len(jsonData))
	return []byte(jsonData), nil
}

func SaveMessageToDB(db *sql.DB, botUsername string, msg *tgbotapi.Message) error {
	if db == nil {
		return nil
	}

	query := `
        INSERT INTO main.messages_log (
            created_at, bot_id, user_id, message_id, chat_id,
            bot_username, message_text, user_name, user_username,
            has_sticker, has_photo, has_document, chat_title
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `

	hasSticker := msg.Sticker != nil
	hasPhoto := msg.Photo != nil && len(msg.Photo) > 0
	hasDocument := msg.Document != nil

	chatTitle := ""
	if msg.Chat.Title != "" {
		chatTitle = msg.Chat.Title
	} else if msg.Chat.UserName != "" {
		chatTitle = "@" + msg.Chat.UserName
	} else {
		chatTitle = msg.Chat.FirstName
	}

	_, err := db.Exec(query,
		time.Now(),
		0,
		msg.From.ID,
		msg.MessageID,
		msg.Chat.ID,
		botUsername,
		msg.Text,
		msg.From.FirstName,
		msg.From.UserName,
		hasSticker,
		hasPhoto,
		hasDocument,
		chatTitle,
	)

	if err != nil {
		log.Printf("❌ Ошибка сохранения в БД: %v", err)
		return err
	}

	log.Printf("💾 Сообщение сохранено в БД (чат: %s)", chatTitle)
	return nil
}
