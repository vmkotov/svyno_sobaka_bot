package mybot

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для рассылки (вынесены в начало файла для удобства)
const (
	// Фразы для рассылки свинособаки дня
	broadcastPhrase1 = "Поздравляем с этим почётным званием! 🐷🐶"
	broadcastPhrase2 = "Это безусловно успех 🎊"

	// Настройки параллелизма
	broadcastMaxWorkers     = 5
	broadcastStartDelay     = 800 * time.Millisecond
	broadcastGoroutineDelay = 50 * time.Millisecond
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

// Обработка одного чата с использованием существующей sendMessage из messages.go
func processChat(bot *tgbotapi.BotAPI, chatID int64, finalName string, wg *sync.WaitGroup, results chan<- string) {
	defer wg.Done()

	chatLog := fmt.Sprintf("Чат %d", chatID)

	// 1. Первое сообщение - используем существующую sendMessage
	sendMessage(bot, chatID,
		"🔍 Идёт сканирование пользователей чата на наличие свинособаки",
		"первое сообщение рассылки")

	// Короткая пауза для эффекта
	time.Sleep(broadcastStartDelay)

	// 2. Второе сообщение - используем вынесенные константы
	msgText := fmt.Sprintf("🎉 СВИНОСОБАКА ДНЯ\n\n"+
		"Сегодня свинособака – это %s\n\n"+
		"%s\n"+
		"%s",
		finalName,
		broadcastPhrase1,
		broadcastPhrase2)

	sendMessage(bot, chatID, msgText, "второе сообщение рассылки")

	results <- fmt.Sprintf("✅ %s: успешно отправлено", chatLog)
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

	// Подготовка данных для параллельной обработки
	type ChatTask struct {
		ChatID    int64
		FinalName string
	}

	var tasks []ChatTask
	chatIDs := []int64{}

	for rows.Next() {
		var chatID int64
		var displayName, userName, userUsername sql.NullString

		if err := rows.Scan(&chatID, &displayName, &userName, &userUsername); err != nil {
			log.Printf("⚠️ Ошибка чтения данных для чата: %v", err)
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

		tasks = append(tasks, ChatTask{ChatID: chatID, FinalName: finalName})
		chatIDs = append(chatIDs, chatID)

		log.Printf("📍 Добавлен чат %d: %s", chatID, finalName)
	}

	// Проверяем ошибки rows
	if err := rows.Err(); err != nil {
		log.Printf("⚠️ Ошибка при итерации rows: %v", err)
	}

	log.Printf("📍 Всего чатов для рассылки: %d", len(tasks))

	// ПАРАЛЛЕЛЬНАЯ ОБРАБОТКА с семафором
	semaphore := make(chan struct{}, broadcastMaxWorkers)
	var wg sync.WaitGroup
	results := make(chan string, len(tasks))

	startTime := time.Now()
	log.Println("🚀 Начинаю параллельную рассылку...")

	// Запускаем воркеры
	for _, task := range tasks {
		wg.Add(1)
		semaphore <- struct{}{} // Занимаем слот в семафоре

		go func(chatID int64, finalName string) {
			defer func() {
				<-semaphore // Освобождаем слот
			}()

			processChat(bot, chatID, finalName, &wg, results)
		}(task.ChatID, task.FinalName)

		// Минимальная задержка между запусками горутин
		time.Sleep(broadcastGoroutineDelay)
	}

	// Ждём завершения всех горутин
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем результаты
	successCount := 0
	failCount := 0

	for result := range results {
		log.Println(result)
		if strings.HasPrefix(result, "✅") {
			successCount++
		} else {
			failCount++
		}
	}

	duration := time.Since(startTime)

	// Статистика
	log.Printf("🎉 Рассылка завершена за %v", duration)
	log.Printf("📈 Статистика:")
	log.Printf("   Всего чатов: %d", len(tasks))
	log.Printf("   Успешно отправлено: %d", successCount)
	log.Printf("   Не удалось отправить: %d", failCount)

	// Рассчитываем примерное время для 100 чатов
	if len(tasks) > 0 {
		timePerChat := duration / time.Duration(len(tasks))
		estimated100 := timePerChat * 100 / time.Duration(broadcastMaxWorkers)
		log.Printf("⏱️  Примерное время для 100 чатов: %v", estimated100)
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
