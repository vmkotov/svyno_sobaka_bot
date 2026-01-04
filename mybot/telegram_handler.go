package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	forwarder "github.com/vmkotov/telegram-forwarder"
)

// HandleMessage - основная функция обработки сообщения
// Принимает бота, сообщение и ID чата для пересылки
// Делает всего две вещи:
//  1. Пересылает сообщение (если нужно)
//  2. Обрабатывает команды (если это команда)
func HandleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, forwardChatID int64) {
	// Логируем что получили
	log.Printf("💬 Сообщение от @%s: %s", msg.From.UserName, msg.Text)

	// 1. ПЕРЕСЫЛАЕМ (если указан chatID)
	if forwardChatID != 0 {
		forwarder.JustForward(bot, msg, forwardChatID)
	}

	// 2. ПРОВЕРЯЕМ КОМАНДЫ
	if msg.IsCommand() {
		handleCommand(bot, msg)
	}
}

// handleCommand - обрабатывает команды
func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		sendStart(bot, msg)
	case "help":
		sendHelp(bot, msg)
	}
}

// sendStart - отправляет приветствие на /start
func sendStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	// Формируем имя
	name := msg.From.FirstName
	if msg.From.UserName != "" {
		name = "@" + msg.From.UserName
	}

	// Текст ответа
	text := fmt.Sprintf(
		"привет, я Свинособака! ты, %s, кстати тоже!\n"+
			"ждём от Грека БТ, ФТ, ТЗ и прочую хуйню.\n"+
			"а пока иди нахуй",
		name)

	// Отправляем
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	if _, err := bot.Send(reply); err != nil {
		log.Printf("❌ Ошибка отправки: %v", err)
	} else {
		log.Printf("✅ Ответил на /start")
	}
}

// sendHelp - отправляет справку на /help
func sendHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "📋 Команды:\n/start - Начать\n/help - Помощь"

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	if _, err := bot.Send(reply); err != nil {
		log.Printf("❌ Ошибка отправки help: %v", err)
	}
}
