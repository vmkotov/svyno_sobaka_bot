package mybot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStartCommand - обработка команды /start
func handleStartCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	// Убрали неиспользуемую переменную name
	text := fmt.Sprintf(
		"Привет! Я бот-свинособака 🐷🐶\n"+
			"Для перезагрузки триггеров из БД нажми кнопку:",
	)

	// Создаем сообщение
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	
	// Добавляем inline-кнопку
	refreshButton := tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить триггеры", "refresh_triggers")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(refreshButton),
	)
	reply.ReplyMarkup = inlineKeyboard

	// Отправляем сообщение с кнопкой
	if _, err := bot.Send(reply); err != nil {
		log.Printf("❌ Ошибка отправки стартового сообщения: %v", err)
	} else {
		log.Printf("✅ Стартовое сообщение с кнопкой отправлено для @%s", msg.From.UserName)
	}
}

// handleHelpCommand - обработка команды /help
func handleHelpCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	text := "📋 Команды:\n/start - Начать\n/help - Помощь\n/refresh_me - Обновить триггеры из БД"
	sendMessage(bot, msg.Chat.ID, text, "справка")
}

// sendMessage - общая функция отправки сообщений
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text, context string) {
	reply := tgbotapi.NewMessage(chatID, text)

	if _, err := bot.Send(reply); err != nil {
		log.Printf("❌ Ошибка отправки %s: %v", context, err)
	} else {
		log.Printf("✅ Отправлен %s", context)
	}
}
