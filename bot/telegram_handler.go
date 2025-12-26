package bot

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramHandler обрабатывает вебхуки от Telegram
type TelegramHandler struct {
	bot *tgbotapi.BotAPI
}

// NewTelegramHandler создает новый обработчик Telegram
func NewTelegramHandler(bot *tgbotapi.BotAPI) *TelegramHandler {
	return &TelegramHandler{
		bot: bot,
	}
}

// HandleWebhook обрабатывает вебхук от Telegram
func (th *TelegramHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("📨 Received webhook (%d bytes)", len(body))
	
	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		log.Printf("❌ Error unmarshaling update: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Обработка сообщения
	if update.Message != nil {
		th.processMessage(&update)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processMessage обрабатывает сообщение
func (th *TelegramHandler) processMessage(update *tgbotapi.Update) {
	msg := update.Message
	log.Printf("💬 Message from @%s: %s", msg.From.UserName, msg.Text)
	
	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			reply := tgbotapi.NewMessage(msg.Chat.ID, 
				"👋 Привет! Я тестовый бот v4.\n" +
				"Использую подход работающего бота!")
			th.bot.Send(reply)
			log.Printf("✅ Sent response to /start")
		case "help":
			reply := tgbotapi.NewMessage(msg.Chat.ID, 
				"📋 Доступные команды:\n" +
				"/start - Начать работу\n" +
				"/help - Помощь")
			th.bot.Send(reply)
		}
	}
}
// Another trigger for auto-deploy
