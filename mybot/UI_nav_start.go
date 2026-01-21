// ============================================================================
// ФАЙЛ: ui_commands.go
// Обработка команд бота: /start, /help, /refresh_me
// ============================================================================
package mybot

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCommand - определяет команду и вызывает соответствующий обработчик
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) {
	switch msg.Command() {
	case "start":
		HandleStartCommand(bot, msg)
	case "help":
		HandleHelpCommand(bot, msg)
	case "refresh_me": // ← команда работает для всех
		HandleRefreshMeCommand(bot, msg, db)
		// Можно добавить другие команды
	}
}

// HandleStartCommand - обработка команды /start
// Теперь разделяет поведение для админов и обычных пользователей
func HandleStartCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	log.Printf("🚀 Команда /start от @%s (ID: %d)", 
		msg.From.UserName, msg.From.ID)

	// Проверяем, является ли пользователь администратором
	if isAdmin(msg.From.ID) {
		log.Printf("👑 Админский /start для @%s", msg.From.UserName)
		SendAdminMainMenu(bot, msg.Chat.ID)
	} else {
		log.Printf("👤 Обычный /start для @%s", msg.From.UserName)
		SendUserMainMenu(bot, msg.Chat.ID)
	}
}

// HandleHelpCommand - обработка команды /help
func HandleHelpCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "📋 Команды:\n" +
		"/start - Начать\n" +
		"/help - Помощь\n" +
		""
	SendMessage(bot, msg.Chat.ID, text, "справка")
}

// HandleRefreshMeCommand - обработка команды /refresh_me
// Работает для всех пользователей (админов и обычных)
func HandleRefreshMeCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) {
	log.Printf("🔄 Команда /refresh_me от @%s (ID: %d)", 
		msg.From.UserName, msg.From.ID)

	// Проверяем подключение к БД
	if db == nil {
		log.Println("⚠️ БД не подключена, не могу обновить триггеры")
		SendMessage(bot, msg.Chat.ID, "❌ БД не подключена", "ошибка")
		return
	}

	// Загружаем конфигурацию
	if err := LoadTriggerConfig(db); err != nil {
		log.Printf("❌ Ошибка загрузки триггеров: %v", err)
		SendMessage(bot, msg.Chat.ID, "❌ Ошибка обновления триггеров", "ошибка")
		return
	}

	// Получаем загруженную конфигурацию
	config := GetTriggerConfig()
	if config == nil || len(config) == 0 {
		log.Println("⚠️ Конфигурация триггеров пуста после загрузки")
		SendMessage(bot, msg.Chat.ID, "✅ Триггеры обновлены!\n⚠️ Но список пуст", "refresh_me")
		return
	}

	log.Println("✅ Триггеры перезагружены из БД")

	// 1. Отправляем сообщение об успехе
	SendMessage(bot, msg.Chat.ID, "✅ Триггеры обновлены!", "refresh_me")

	// 2. Формируем статистику и список
	statsText := FormatTriggerStats(config)
	listText := FormatTriggersList(config)

	// 3. Отправляем статистику
	SendMessage(bot, msg.Chat.ID, statsText, "статистика триггеров")

	// 4. Отправляем список (разбиваем если длинный)
	maxMsgLength := 4000 // Оставляем запас от 4096
	listParts := SplitLongMessage(listText, maxMsgLength)

	for i, part := range listParts {
		context := "список триггеров"
		if len(listParts) > 1 {
			context = fmt.Sprintf("список триггеров (часть %d/%d)", i+1, len(listParts))
		}
		SendMessage(bot, msg.Chat.ID, part, context)
	}
}

// SendUserMainMenu отправляет меню для обычных пользователей
func SendUserMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Привет! Я бот-свинособака 🐷🐶\n" +
		"Я реагирую на сообщения в чатах.\n\n" +
		"Используйте /help для списка команд."

	// Кнопка "О боте"
	aboutButton := tgbotapi.NewInlineKeyboardButtonData(
		"❓ О боте", 
		"menu:about",
	)
	
	// Кнопка "СвиноАдминка" (будет проверять права)
	adminButton := tgbotapi.NewInlineKeyboardButtonData(
		"🐷 СвиноАдминка", 
		"admin:menu",
	)
	
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(aboutButton, adminButton),
	)
	
	// Отправляем сообщение
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = inlineKeyboard
	
	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки пользовательского меню: %v", err)
	} else {
		log.Printf("✅ Пользовательское меню отправлено в чат %d", chatID)
	}
}

// SendAdminMainMenu отправляет главное меню админки "СвиноАдминка"
func SendAdminMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	// Используем ту же функцию что и для пользователей
	SendUserMainMenu(bot, chatID)
	log.Printf("👑 Админское меню отправлено в чат %d", chatID)
}
