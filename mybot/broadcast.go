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
	log.Println("🔄 Вызываем процедуру выбора свинособаки дня...")
	_, err := db.Exec(`CALL svyno_sobaka_bot.proc_svyno_sobaka_of_the_day()`)
	if err != nil {
		log.Printf("❌ Ошибка вызова процедуры: %v", err)
		return err
	}

	log.Println("✅ Таблица заполнена, начинаю рассылку...")

	// 2. Берём сегодняшние записи из таблицы svyno_sobaka_of_the_day
	rows, err := db.Query(`
        SELECT 
            ss.chat_id,
            COALESCE(ss.user_username, ss.user_name, 'Анонимный пользователь') as display_name,
            ss.user_name,
            ss.user_username
        FROM svyno_sobaka_bot.svyno_sobaka_of_the_day ss
        WHERE 1=1
            AND ss.dt_insert::date = CURRENT_DATE
            AND (ss.user_username IS NOT NULL OR ss.user_name IS NOT NULL)
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
		var displayName, userName, userUsername sql.NullString

		if err := rows.Scan(&chatID, &displayName, &userName, &userUsername); err != nil {
			log.Printf("⚠️ Ошибка чтения данных: %v", err)
			continue
		}

		// Определяем как показывать имя
		var finalDisplayName string
		if userUsername.Valid && userUsername.String != "" {
			finalDisplayName = "@" + userUsername.String
		} else if userName.Valid && userName.String != "" {
			finalDisplayName = userName.String
		} else {
			finalDisplayName = "Анонимный пользователь"
		}

		// 1. Первое сообщение - "Идёт сканирование..."
		msg1 := tgbotapi.NewMessage(chatID, "🔍 *Идёт сканирование пользователей чата на наличие свинособаки*")
		msg1.ParseMode = "Markdown"

		if _, err := bot.Send(msg1); err != nil {
			log.Printf("⚠️ Не удалось отправить первое сообщение в чат %d: %v", chatID, err)
			continue
		}

		// Пауза 3 секунды
		time.Sleep(10 * time.Second)

		// 2. Второе сообщение - результат
		msg2 := tgbotapi.NewMessage(chatID,
			"🎉 *СВИНОСОБАКА ДНЯ*\n\n"+
				"Сегодня свинособака – это *"+finalDisplayName+"*\n\n"+
				"Поздравляем с этим почётным званием! 🐷🐶\n"+
				"Не забывайте быть активными в чате!\n\n"+
				"А пока иди нахуй! 🎊")
		msg2.ParseMode = "Markdown"

		if _, err := bot.Send(msg2); err != nil {
			log.Printf("⚠️ Не удалось отправить второе сообщение в чат %d: %v", chatID, err)
			continue
		}

		sentCount++

		// Логируем прогресс
		if sentCount%10 == 0 {
			log.Printf("📤 Отправлено %d сообщений", sentCount)
		}

		// Пауза между чатами
		time.Sleep(500 * time.Millisecond)
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
