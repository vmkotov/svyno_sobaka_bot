package ui

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendMainMenu отправляет главное меню бота
func SendMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	text := fmt.Sprintf(
		"Привет! Я бот-свинособака 🐷🐶\n" +
		"Выберите действие:",
	)

	// Создаем сообщение
	reply := tgbotapi.NewMessage(chatID, text)
	
	// Создаем inline-клавиатуру с двумя кнопками
	refreshButton := tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить триггеры", "refresh:triggers")
	showButton := tgbotapi.NewInlineKeyboardButtonData("📋 Триггеры", "triggers:list")
	
	// Две кнопки в один ряд
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(refreshButton, showButton),
	)
	reply.ReplyMarkup = inlineKeyboard

	// Отправляем сообщение с кнопками
	if _, err := bot.Send(reply); err != nil {
		log.Printf("❌ Ошибка отправки главного меню: %v", err)
	} else {
		log.Printf("✅ Главное меню отправлено в чат %d", chatID)
	}
}

// EditMessageToMainMenu редактирует существующее сообщение, заменяя его главным меню
func EditMessageToMainMenu(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	text := fmt.Sprintf(
		"Привет! Я бот-свинособака 🐷🐶\n" +
		"Выберите действие:",
	)

	// Создаем inline-клавиатуру
	refreshButton := tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить триггеры", "refresh:triggers")
	showButton := tgbotapi.NewInlineKeyboardButtonData("📋 Триггеры", "triggers:list")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(refreshButton, showButton),
	)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		chatID,
		messageID,
		text,
		inlineKeyboard,
	)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка редактирования в главное меню: %v", err)
		// Если не удалось отредактировать, отправляем новое
		SendMainMenu(bot, chatID)
	} else {
		log.Printf("✅ Сообщение отредактировано в главное меню")
	}
}
