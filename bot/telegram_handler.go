package bot

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramHandler обрабатывает вебхуки от Telegram
type TelegramHandler struct {
    bot       *tgbotapi.BotAPI
    forwarder *MessageForwarder
}

// NewTelegramHandler создает новый обработчик Telegram
func NewTelegramHandler(bot *tgbotapi.BotAPI, forwarder *MessageForwarder) *TelegramHandler {
    return &TelegramHandler{
        bot:       bot,
        forwarder: forwarder,
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
    
    // =========================================
    // ПЕРЕСЫЛКА СООБЩЕНИЙ
    // =========================================
    if th.forwarder != nil {
        th.forwarder.Forward(msg)
    }
    
    if msg.IsCommand() {
        switch msg.Command() {
        case "start":
            userName := msg.From.FirstName
            if msg.From.UserName != "" {
                userName = "@" + msg.From.UserName
            }
            
            replyText := fmt.Sprintf("привет, я Свинособака. ты, %s, кстати тоже!\n" +
                "ждём от Грека БТ, ФТ, ТЗ и прочую хуйню.\n" +
                "а пока иди нахуй", userName)
            
            reply := tgbotapi.NewMessage(msg.Chat.ID, replyText)
            _, err := th.bot.Send(reply)
            if err != nil {
                log.Printf("❌ Error sending message: %v", err)
            }
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
