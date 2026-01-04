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

// TelegramHandler - обработчик вебхуков от Telegram API.
// Основные функции:
// 1. Прием и парсинг обновлений от Telegram через вебхуки
// 2. Пересылка всех входящих сообщений в указанный чат через JustForward
// 3. Обработка команд /start и /help с шутливыми ответами
//
// Структура не сохраняет состояние между запросами. Для пересылки сообщений
// используется функция JustForward из модуля forwarder.
type TelegramHandler struct {
	bot           *tgbotapi.BotAPI // Основной клиент для работы с Telegram API
	forwardChatID int64            // ID чата для пересылки сообщений (0 если отключено)
}

// NewTelegramHandler создает и возвращает новый экземпляр TelegramHandler.
// Параметры:
//   - bot: инициализированный клиент Telegram Bot API
//   - forwardChatID: ID чата для пересылки сообщений (0 чтобы отключить пересылку)
//
// Если forwardChatID не равен 0, все входящие сообщения будут автоматически
// пересылаться в указанный чат с помощью функции forwarder.JustForward.
func NewTelegramHandler(bot *tgbotapi.BotAPI, forwardChatID int64) *TelegramHandler {
	return &TelegramHandler{
		bot:           bot,
		forwardChatID: forwardChatID,
	}
}

// HandleWebhook - обработчик HTTP запросов от Telegram Webhook.
// Этот метод должен быть зарегистрирован как обработчик для пути,
// на который Telegram отправляет обновления (например, "/webhook").
//
// Telegram отправляет обновления в формате JSON методом POST.
// После успешной обработки метод возвращает HTTP 200 OK.
// При ошибках валидации или парсинга возвращаются соответствующие HTTP коды ошибок.
func (th *TelegramHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Проверяем что это POST запрос - Telegram отправляет обновления только методом POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Читаем тело запроса полностью
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

	// Обрабатываем сообщение если оно есть в обновлении
	if update.Message != nil {
		th.processMessage(&update)
	}

	// Отвечаем 200 OK - подтверждаем Telegram, что обновление получено
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processMessage - обрабатывает входящее сообщение из обновления.
// Выполняет три основные действия:
//  1. Логирует сообщение для отладки
//  2. Пересылает сообщение через JustForward (если forwardChatID настроен)
//  3. Обрабатывает команды, если сообщение является командой
func (th *TelegramHandler) processMessage(update *tgbotapi.Update) {
	msg := update.Message

	// Логируем полученное сообщение
	log.Printf("💬 Сообщение от @%s: %s",
		msg.From.UserName,
		msg.Text)

	// =========================================
	// ПЕРЕСЫЛКА СООБЩЕНИЙ ЧЕРЕЗ JUSTFORWARD
	// =========================================
	// Используем новую простую функцию JustForward вместо объекта MessageForwarder
	if th.forwardChatID != 0 {
		forwarder.JustForward(th.bot, msg, th.forwardChatID)
	}

	// Обрабатываем команды (сообщения, начинающиеся с "/")
	if msg.IsCommand() {
		th.handleCommand(msg)
	}
}

// handleCommand - определяет тип команды и вызывает соответствующий обработчик.
func (th *TelegramHandler) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		th.handleStartCommand(msg)
	case "help":
		th.handleHelpCommand(msg)
		// Другие команды не обрабатываются
	}
}

// handleStartCommand - обрабатывает команду /start.
// Отправляет пользователю шутливое/оскорбительное приветствие.
func (th *TelegramHandler) handleStartCommand(msg *tgbotapi.Message) {
	userName := msg.From.FirstName
	if msg.From.UserName != "" {
		userName = "@" + msg.From.UserName
	}

	replyText := fmt.Sprintf(
		"привет, я Свинособака. ты, %s, кстати тоже!\n"+
			"ждём от Грека БТ, ФТ, ТЗ и прочую хуйню.\n"+
			"а пока иди нахуй",
		userName)

	reply := tgbotapi.NewMessage(msg.Chat.ID, replyText)
	_, err := th.bot.Send(reply)
	if err != nil {
		log.Printf("❌ Ошибка отправки сообщения: %v", err)
	} else {
		log.Printf("✅ Отправлен ответ на /start")
	}
}

// handleHelpCommand - обрабатывает команду /help.
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
