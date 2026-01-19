// ФАЙЛ: mybot/db_detailed.go
package mybot

import (
	"database/sql"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SaveMessageDetailed - сохраняет сообщение детально во все таблицы
// Каждая вставка коммитится отдельно (авто-коммит)
func SaveMessageDetailed(db *sql.DB, botUser *tgbotapi.User, msg *tgbotapi.Message) error {
	if db == nil {
		return nil
	}

	log.Printf("💾 Детальное сохранение сообщения %d от @%s",
		msg.MessageID, msg.From.UserName)

	startTime := time.Now()
	defer func() {
		log.Printf("⏱️ Детальное сохранение заняло: %v", time.Since(startTime))
	}()

	// ===========================================
	// 1. ВСТАВКА В БД (с авто-коммитами)
	// ===========================================

	// 1.1. Пользователь (отправитель)
	if err := upsertUser(db, botUser.ID, msg.From); err != nil {
		log.Printf("⚠️ Ошибка upsert_user: %v (продолжаем)", err)
	}

	// 1.2. Чат
	if err := upsertChat(db, msg.Chat); err != nil {
		log.Printf("⚠️ Ошибка upsert_chat: %v (продолжаем)", err)
	}

	// 1.3. Основное сообщение
	if err := insertMessage(db, botUser.ID, msg); err != nil {
		log.Printf("⚠️ Ошибка insert_message: %v (продолжаем)", err)
	}

	// 1.4. Медиафайлы
	if err := insertMedia(db, msg); err != nil {
		log.Printf("⚠️ Ошибка insert_media: %v (продолжаем)", err)
	}

	// 1.5. Ответ на сообщение (reply)
	if msg.ReplyToMessage != nil {
		if err := insertReplyReference(db, msg); err != nil {
			log.Printf("⚠️ Ошибка insert_reply: %v (продолжаем)", err)
		}
	}

	// 1.6. Пересылка (forward)
	if msg.ForwardFrom != nil || msg.ForwardFromChat != nil {
		if err := insertForwardReference(db, msg); err != nil {
			log.Printf("⚠️ Ошибка insert_forward: %v (продолжаем)", err)
		}
	}

	// 1.7. Упоминания пользователей
	if err := insertMentions(db, msg); err != nil {
		log.Printf("⚠️ Ошибка insert_mentions: %v (продолжаем)", err)
	}

	log.Printf("✅ Детальное сохранение завершено для сообщения %d", msg.MessageID)
	return nil
}

// ===========================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ===========================================

// safeString возвращает безопасную строку (не nil)
func safeString(s string) string {
	if s == "" {
		return ""
	}
	return s
}

// upsertUser - вызов процедуры upsert_user
func upsertUser(db *sql.DB, botID int64, from *tgbotapi.User) error {
	if from == nil {
		return nil
	}

	// Сначала сохраняем отправителя
	_, err := db.Exec(
		`CALL svyno_sobaka_bot.upsert_user($1, $2, $3, $4, $5, $6)`,
		from.ID,
		from.IsBot,
		safeString(from.FirstName),
		safeString(from.LastName),
		safeString(from.UserName),
		safeString(from.LanguageCode),
	)

	return err
}

// upsertChat - вызов процедуры upsert_chat
func upsertChat(db *sql.DB, chat *tgbotapi.Chat) error {
	if chat == nil {
		return nil
	}

	_, err := db.Exec(
		`CALL svyno_sobaka_bot.upsert_chat($1, $2, $3, $4, $5, $6, $7)`,
		chat.ID,
		chat.Type,
		safeString(chat.Title),
		safeString(chat.UserName),
		safeString(chat.FirstName),
		safeString(chat.LastName),
		"", // description (пока пустой)
	)

	return err
}

// insertMessage - вызов процедуры insert_message (ИСПРАВЛЕННЫЙ!)
func insertMessage(db *sql.DB, botID int64, msg *tgbotapi.Message) error {
	messageText := msg.Text
	caption := msg.Caption

	// Если нет текста, но есть подпись к медиа
	if messageText == "" && caption != "" {
		messageText = caption
		caption = ""
	}

	// Дата сообщения из Unix timestamp
	messageDate := time.Unix(int64(msg.Date), 0)

	// ВАЖНО: Правильный порядок параметров согласно сигнатуре процедуры:
	// insert_message(
	//   n_chat_id bigint,           -- 1
	//   n_message_id integer,       -- 2
	//   d_message_date timestamptz, -- 3
	//   n_user_id bigint,           -- 4 ← ДОЛЖЕН быть user_id!
	//   v_message_text text,        -- 5
	//   v_caption text,             -- 6
	//   b_is_bot boolean,           -- 7
	//   n_telegram_update_id bigint -- 8
	// )
	_, err := db.Exec(
		`CALL svyno_sobaka_bot.insert_message($1, $2, $3, $4, $5, $6, $7, $8)`,
		msg.Chat.ID,    // 1. n_chat_id (bigint)
		msg.MessageID,  // 2. n_message_id (integer)
		messageDate,    // 3. d_message_date (timestamptz)
		msg.From.ID,    // 4. n_user_id (bigint) ← ВАЖНО! РАНЬШЕ БЫЛ ТЕКСТ
		messageText,    // 5. v_message_text (text)
		caption,        // 6. v_caption (text)
		msg.From.IsBot, // 7. b_is_bot (boolean)
		nil,            // 8. n_telegram_update_id (bigint)
	)

	return err
}

// insertMedia - обработка медиафайлов
func insertMedia(db *sql.DB, msg *tgbotapi.Message) error {
	var err error

	// Фото (может быть несколько)
	if msg.Photo != nil && len(msg.Photo) > 0 {
		// Берем самую большую фото (последнюю в массиве)
		photo := msg.Photo[len(msg.Photo)-1]
		_, err = db.Exec(
			`CALL svyno_sobaka_bot.insert_media($1, $2, $3, $4, $5, $6, $7, $8)`,
			msg.Chat.ID,
			msg.MessageID,
			"photo",
			photo.FileID,
			photo.FileUniqueID,
			photo.FileSize,
			photo.Width,
			photo.Height,
		)
		if err != nil {
			return err
		}
	}

	// Документ
	if msg.Document != nil {
		_, err = db.Exec(
			`CALL svyno_sobaka_bot.insert_media($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			msg.Chat.ID,
			msg.MessageID,
			"document",
			msg.Document.FileID,
			msg.Document.FileUniqueID,
			msg.Document.FileSize,
			nil, // width
			nil, // height
			nil, // duration
			msg.Document.MimeType,
			msg.Document.FileName,
		)
		if err != nil {
			return err
		}
	}

	// Стикер
	if msg.Sticker != nil {
		_, err = db.Exec(
			`CALL svyno_sobaka_bot.insert_media($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			msg.Chat.ID,
			msg.MessageID,
			"sticker",
			msg.Sticker.FileID,
			msg.Sticker.FileUniqueID,
			msg.Sticker.FileSize,
			msg.Sticker.Width,
			msg.Sticker.Height,
			nil, // duration
			"",  // mime_type
			"",  // file_name
			msg.Sticker.Emoji,
		)
		if err != nil {
			return err
		}
	}

	// Видео
	if msg.Video != nil {
		_, err = db.Exec(
			`CALL svyno_sobaka_bot.insert_media($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			msg.Chat.ID,
			msg.MessageID,
			"video",
			msg.Video.FileID,
			msg.Video.FileUniqueID,
			msg.Video.FileSize,
			msg.Video.Width,
			msg.Video.Height,
			msg.Video.Duration,
			msg.Video.MimeType,
		)
		if err != nil {
			return err
		}
	}

	// Аудио
	if msg.Audio != nil {
		_, err = db.Exec(
			`CALL svyno_sobaka_bot.insert_media($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			msg.Chat.ID,
			msg.MessageID,
			"audio",
			msg.Audio.FileID,
			msg.Audio.FileUniqueID,
			msg.Audio.FileSize,
			nil, // width
			nil, // height
			msg.Audio.Duration,
			msg.Audio.MimeType,
			"", // file_name
			"", // emoji
			msg.Audio.Performer,
			msg.Audio.Title,
		)
		if err != nil {
			return err
		}
	}

	// Голосовое сообщение
	if msg.Voice != nil {
		_, err = db.Exec(
			`CALL svyno_sobaka_bot.insert_media($1, $2, $3, $4, $5, $6, $7)`,
			msg.Chat.ID,
			msg.MessageID,
			"voice",
			msg.Voice.FileID,
			msg.Voice.FileUniqueID,
			msg.Voice.FileSize,
			nil, // width
			nil, // height
			msg.Voice.Duration,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// insertReplyReference - обработка reply
func insertReplyReference(db *sql.DB, msg *tgbotapi.Message) error {
	if msg.ReplyToMessage == nil {
		return nil
	}

	// Сохраняем пользователя, на чье сообщение ответили
	if msg.ReplyToMessage.From != nil {
		upsertUser(db, 0, msg.ReplyToMessage.From)
	}

	_, err := db.Exec(
		`CALL svyno_sobaka_bot.insert_reference($1, $2, $3, $4, $5, $6)`,
		msg.Chat.ID,
		msg.MessageID,
		"reply",
		msg.Chat.ID,
		msg.ReplyToMessage.MessageID,
		msg.ReplyToMessage.From.ID,
	)

	return err
}

// insertForwardReference - обработка forward
func insertForwardReference(db *sql.DB, msg *tgbotapi.Message) error {
	// Пока упрощенная версия, без всех полей
	var forwardDate *time.Time
	if msg.ForwardDate > 0 {
		date := time.Unix(int64(msg.ForwardDate), 0)
		forwardDate = &date
	}

	_, err := db.Exec(
		`CALL svyno_sobaka_bot.insert_reference($1, $2, $3, $4, $5, $6, $7)`,
		msg.Chat.ID,
		msg.MessageID,
		"forward",
		nil, // referenced_chat_id
		nil, // referenced_message_id
		nil, // referenced_user_id
		forwardDate,
	)

	return err
}

// insertMentions - обработка упоминаний (пока заглушка)
func insertMentions(db *sql.DB, msg *tgbotapi.Message) error {
	// TODO: Реализовать парсинг упоминаний из текста
	// Пока просто возвращаем nil
	return nil
}
