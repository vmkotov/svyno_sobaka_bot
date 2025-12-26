package main

import (
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"svyno_sobaka_bot/bot"
)

func main() {
	log.Println("🚀 Запуск тестового бота v4 (аналог работающего)...")

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

	// Создаем обработчик Telegram как в работающем боте
	telegramHandler := bot.NewTelegramHandler(botAPI)

	// Настраиваем HTTP роутер
	http.HandleFunc("/", telegramHandler.HandleWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌐 Сервер на порту %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
