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
			if err := SendSvynoSobakaBroadcast(bot, db, "svyno_sobaka_bot"); err != nil {
				log.Printf("❌ Ошибка рассылки: %v", err)
			}
		}()

		// 3. Отвечаем сразу
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Svyno sobaka broadcast started"))
	}
}

// SendSvynoSobakaBroadcast выполняет рассылку с выбором свинособаки дня
func SendSvynoSobakaBroadcast(bot *tgbotapi.BotAPI, db *sql.DB, botUsername string) error {
	if db == nil {
		log.Println("ℹ️ БД не настроена, пропускаем рассылку")
		return nil
	}

	log.Println("📢 Начинаю рассылку свинособаки дня...")

	// 1. Сначала вызываем процедуру для заполнения таблицы
	_, err := db.Exec(`CALL svyno_sobaka_bot.proc_svyno_sobaka_of_the_day()`)
	if err != nil {
		log.Printf("❌ Ошибка вызова процедуры: %v", err)
		return err
	}

	log.Println("✅ Таблица заполнена, начинаю рассылку...")

	// 2. Берём сегодняшние записи из таблицы svyno_sobaka_of_the_day
	rows, err := db.Query(`
        SELECT 
         ss.chat_id
         , ss.user_username
         , ss.user_name
        FROM svyno_sobaka_bot.svyno_sobaka_of_the_day ss
        WHERE 1=1
         AND ss.dt_insert::date = CURRENT_DATE
         AND ss.user_username IS NOT NULL
        ORDER BY ss.chat_id
    `)

	if err != nil {
		log.Printf("❌ Ошибка запроса к таблице свинособак: %v", err)
		return err
	}
	defer rows.Close()

	// 3. Рассылаем по каждому чату
	sentCount := 0
	for rows.Next() {
		var chatID int64
		var username, name sql.NullString

		if err := rows.Scan(&chatID, &username, &name); err != nil {
			log.Printf("⚠️ Ошибка чтения данных: %v", err)
			continue
		}

		// Формируем сообщение
		var messageText string
		var displayName string

		// Выбираем что показывать: username или name
		if username.Valid && username.String != "" {
			displayName = "@" + username.String
		} else if name.Valid && name.String != "" {
			displayName = name.String
		} else {
			displayName = "Анонимный пользователь"
		}

		messageText =
			"🎉 *СВИНОСОБАКА ДНЯ* 🎉\n\n" +
				"Сегодняшняя свинособака дня: " + displayName + "\n\n" +
				"Поздравляем! 🐷🐶\n" +
				"Это почётное звание за активность в чате!\n\n" +
				"А пока иди нахуй! 🎊"

		msg := tgbotapi.NewMessage(chatID, messageText)
		msg.ParseMode = "Markdown"

		if _, err := bot.Send(msg); err != nil {
			log.Printf("⚠️ Не удалось отправить в чат %d: %v", chatID, err)
			continue
		}

		sentCount++

		// Логируем прогресс
		if sentCount%10 == 0 {
			log.Printf("📤 Отправлено %d сообщений", sentCount)
		}

		// Пауза между сообщениями
		time.Sleep(200 * time.Millisecond)
	}

	if err := rows.Err(); err != nil {
		log.Printf("⚠️ Ошибка при итерации по результатам: %v", err)
	}

	log.Printf("🎉 Рассылка завершена. Отправлено: %d сообщений", sentCount)
	return nil
}

// isAuthorized проверяет авторизацию (остаётся без изменений)
func isAuthorized(r *http.Request, secretKey string) bool {
	// 1. Разрешаем локальные запросы
	if strings.HasPrefix(r.RemoteAddr, "127.0.0.1") ||
		strings.HasPrefix(r.RemoteAddr, "[::1]") {
		return true
	}

	// 2. Разрешаем по секретному заголовку
	if r.Header.Get("X-Broadcast-Secret") == secretKey {
		return true
	}

	// 3. Разрешаем по User-Agent Yandex Cloud
	userAgent := strings.ToLower(r.UserAgent())
	if strings.Contains(userAgent, "yandex") ||
		strings.Contains(userAgent, "cloud") {
		return true
	}

	return false
}
