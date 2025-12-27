package bot

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    forwarder "github.com/vmkotov/telegram-forwarder"
)

// TelegramHandler обрабатывает вебхуки от Telegram
type TelegramHandler struct {
    bot       *tgbotapi.BotAPI
    forwarder *forwarder.MessageForwarder
}

// NewTelegramHandler создает новый обработчик Telegram
func NewTelegramHandler(bot *tgbotapi.BotAPI, forwarder *forwarder.MessageForwarder) *TelegramHandler {
    return &TelegramHandler{
        bot:       bot,
        forwarder: forwarder,
    }
}

// HandleWebhook обрабатывает вебхук от Telegram
func (th *TelegramHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    // Проверяем что это POST запрос
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Читаем тело запроса
    body, err := io.ReadAll(r.Body)
    if err != nil {
        log.Printf("❌ Ошибка чтения тела запроса: %v", err)
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    log.Printf("📨 Получен вебхук (%d байт)", len(body))
    
    // Парсим JSON в структуру Update
    var update tgbotapi.Update
    if err := json.Unmarshal(body, &update); err != nil {
        log.Printf("❌ Ошибка парсинга update: %v", err)
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // Обрабатываем сообщение если оно есть
    if update.Message != nil {
        th.processMessage(&update)
    }

    // Отвечаем 200 OK
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

// processMessage обрабатывает входящее сообщение
func (th *TelegramHandler) processMessage(update *tgbotapi.Update) {
    msg := update.Message
    
    // Логируем полученное сообщение
    log.Printf("💬 Сообщение от @%s: %s", 
        msg.From.UserName, 
        msg.Text)
    
    // =========================================
    // ПЕРЕСЫЛКА СООБЩЕНИЙ В ЦЕЛЕВОЙ ЧАТ
    // =========================================
    if th.forwarder != nil {
        th.forwarder.Forward(msg)
    }
    
    // Обрабатываем команды
    if msg.IsCommand() {
        th.handleCommand(msg)
    }
}

// handleCommand обрабатывает команды бота
func (th *TelegramHandler) handleCommand(msg *tgbotapi.Message) {
    switch msg.Command() {
    case "start":
        th.handleStartCommand(msg)
    case "help":
        th.handleHelpCommand(msg)
    }
}

// handleStartCommand обрабатывает команду /start
func (th *TelegramHandler) handleStartCommand(msg *tgbotapi.Message) {
    // Формируем имя пользователя для приветствия
    userName := msg.From.FirstName
    if msg.From.UserName != "" {
        userName = "@" + msg.From.UserName
    }
    
    // Текст приветствия
    replyText := fmt.Sprintf(
        "привет, я Свинособака. ты, %s, кстати тоже!\n" +
        "ждём от Грека БТ, ФТ, ТЗ и прочую хуйню.\n" +
        "а пока иди нахуй", 
        userName)
    
    // Отправляем ответ
    reply := tgbotapi.NewMessage(msg.Chat.ID, replyText)
    _, err := th.bot.Send(reply)
    if err != nil {
        log.Printf("❌ Ошибка отправки сообщения: %v", err)
    } else {
        log.Printf("✅ Отправлен ответ на /start")
    }
}

// handleHelpCommand обрабатывает команду /help
func (th *TelegramHandler) handleHelpCommand(msg *tgbotapi.Message) {
    replyText := "📋 Доступные команды:\n" +
                 "/start - Начать работу\n" +
                 "/help - Помощь"
    
    reply := tgbotapi.NewMessage(msg.Chat.ID, replyText)
    _, err := th.bot.Send(reply)
    if err != nil {
        log.Printf("❌ Ошибка отправки help: %v", err)
    }
}
