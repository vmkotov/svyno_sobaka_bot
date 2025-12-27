package main

import (
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	forwarder "github.com/vmkotov/telegram-forwarder"
	"github.com/joho/godotenv"

	"svyno_sobaka_bot/bot"
)

func main() {
	// =========================================
	// ИНИЦИАЛИЗАЦИЯ БОТА
	// =========================================
	log.Println("🚀 Запуск тестового бота v4 (аналог работающего)...")

	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ Файл .env не найден: %v", err)
	}

	// Получаем токен бота из переменных окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN не найден")
	}

	// Создаем экземпляр бота
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("❌ Ошибка создания бота: %v", err)
	}

	// Включаем режим отладки
	botAPI.Debug = true
	log.Printf("✅ Авторизован как @%s", botAPI.Self.UserName)

	// =========================================
	// СОЗДАЕМ FORWARDER ИЗ ВНЕШНЕЙ БИБЛИОТЕКИ
	// =========================================
	forwardChatID := int64(-1003677836395) // ID чата для пересылки (хардкод)
	fwd := forwarder.New(botAPI, forwardChatID)
	log.Printf("📍 ID чата для пересылки: %d", forwardChatID)

	// Создаем обработчик Telegram с forwarder
	telegramHandler := bot.NewTelegramHandler(botAPI, fwd)

	// =========================================
	// НАСТРАИВАЕМ HTTP СЕРВЕР
	// =========================================
	http.HandleFunc("/", telegramHandler.HandleWebhook)

	// Получаем порт из переменных окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Запускаем HTTP сервер
	log.Printf("🌐 Сервер запущен на порту %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
// Auto-deploy trigger пятница, 26 декабря 2025 г. 22:15:14 (MSK)
