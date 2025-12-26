package main

import (
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("🚀 Запуск нового тестового бота...")

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

	log.Printf("✅ Авторизован как @%s", botAPI.Self.UserName)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		update, err := botAPI.HandleUpdate(r)
		if err != nil {
			log.Printf("❌ Ошибка вебхука: %v", err)
			return
		}

		if update.Message != nil && update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, 
					"👋 Привет! Я новый тестовый бот.\n"+
					"Пока умею только отвечать на /start")
				botAPI.Send(msg)
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, 
					"📋 Доступные команды:\n"+
					"/start - Начать работу\n"+
					"/help - Помощь")
				botAPI.Send(msg)
			}
		}
		w.WriteHeader(http.StatusOK)
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
// Trigger deployment
// Test commit to trigger deployment
