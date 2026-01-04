package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	mybot "svyno_sobaka_bot/mybot" // Теперь понятнее!
)

func main() {
	log.Println("🚀 Запуск простого бота...")

	// Загружаем настройки
	godotenv.Load()

	// Создаём бота
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	botAPI, _ := tgbotapi.NewBotAPI(token)
	log.Printf("✅ Бот: @%s", botAPI.Self.UserName)

	// Куда пересылать сообщения
	forwardChatID := int64(-1003677836395)

	// Настраиваем обработчик HTTP
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, botAPI, forwardChatID)
	})

	// Запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("🌐 Сервер на порту %s", port)
	http.ListenAndServe(":"+port, nil)
}

// handleRequest - обрабатывает один HTTP запрос
func handleRequest(w http.ResponseWriter, r *http.Request, bot *tgbotapi.BotAPI, forwardChatID int64) {
	// Только POST запросы
	if r.Method != "POST" {
		http.Error(w, "Нужен POST", http.StatusMethodNotAllowed)
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Ошибка чтения: %v", err)
		http.Error(w, "Ошибка чтения", http.StatusBadRequest)
		return
	}

	// Парсим JSON от Telegram
	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		log.Printf("❌ Ошибка парсинга JSON: %v", err)
		http.Error(w, "Неправильный JSON", http.StatusBadRequest)
		return
	}

	// Если есть сообщение - обрабатываем
	if update.Message != nil {
		// Вызываем функцию из нашего пакета
		mybot.HandleMessage(bot, update.Message, forwardChatID)
	}

	// Отвечаем "OK"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
