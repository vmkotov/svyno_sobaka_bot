package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("🚀 Запуск тестового бота...")

	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ Файл .env не найден: %v", err)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN не найден")
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("❌ Ошибка создания бота: %v", err)
	}

	botAPI.Debug = true
	log.Printf("✅ Авторизован как @%s", botAPI.Self.UserName)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("📨 Получен запрос: %s %s", r.Method, r.URL.Path)
		log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))
		
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("❌ Ошибка чтения тела: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		
		log.Printf("📦 Тело запроса (%d байт): %s", len(body), string(body))
		
		var update tgbotapi.Update
		if err := json.Unmarshal(body, &update); err != nil {
			log.Printf("❌ Ошибка парсинга JSON: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		
		// Обработка сообщения
		if update.Message != nil {
			log.Printf("💬 Сообщение от @%s: %s", 
				update.Message.From.UserName, 
				update.Message.Text)
			
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, 
						"👋 Привет! Я тестовый бот v2.\n" +
						"Вебхуки работают!")
					botAPI.Send(msg)
					log.Printf("✅ Отправлен ответ на /start")
				}
			}
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌐 Сервер на порту %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
