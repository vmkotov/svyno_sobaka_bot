package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"svyno_sobaka_bot/mybot"
)

func main() {
	log.Println("🚀 Запуск бота с БД и рассылкой...")
	godotenv.Load()

	// 1. Бот
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN не найден")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("❌ Ошибка создания бота: %v", err)
	}

	log.Printf("✅ Бот: @%s", bot.Self.UserName)

	// 2. БД (если есть настройки)
	var db *sql.DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Printf("⚠️ Не удалось подключиться к БД: %v", err)
			db = nil
		} else {
			// Проверяем подключение
			if err := db.Ping(); err != nil {
				log.Printf("⚠️ БД недоступна: %v", err)
				db = nil
			} else {
				defer db.Close()
				log.Println("✅ Подключено к PostgreSQL")
			}
		}
	}

	// 3. ID для пересылки
	forwardChatID := int64(-1003677836395)

	// 4. Секретный ключ для рассылки
	broadcastSecret := os.Getenv("BROADCAST_SECRET")
	if broadcastSecret == "" {
		broadcastSecret = "change-me-in-production"
		log.Println("⚠️ Используется дефолтный BROADCAST_SECRET")
	}

	// 5. Настраиваем HTTP обработчики
	// Основной вебхук от Telegram
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleWebhook(w, r, bot, forwardChatID, db)
	})

	// Эндпоинт для рассылки
	broadcastHandler := mybot.SetupBroadcastHandler(bot, db, broadcastSecret)
	http.HandleFunc("/admin/broadcast", broadcastHandler)

	// 6. Стартуем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("🌐 Сервер запущен на порту %s", port)
	log.Printf("📢 Эндпоинт рассылки: http://localhost:%s/admin/broadcast", port)
	log.Println("📝 Заголовок: X-Broadcast-Secret: " + broadcastSecret)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// handleWebhook обрабатывает вебхук от Telegram
func handleWebhook(w http.ResponseWriter, r *http.Request, bot *tgbotapi.BotAPI,
	forwardChatID int64, db *sql.DB) {

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Ошибка чтения: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		log.Printf("❌ Ошибка парсинга JSON: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if update.Message != nil {
		mybot.HandleMessage(bot, update.Message, forwardChatID, db, bot.Self.UserName)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
