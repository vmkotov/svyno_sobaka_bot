package mybot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для меню триггеров
const (
	triggersPerPage = 10 // Триггеров на страницу
	maxNameLength   = 25 // Максимальная длина названия в кнопке
)

// generateTriggersMenu создает меню с триггерами для указанной страницы
// Возвращает текст сообщения и inline-клавиатуру
func generateTriggersMenu(page int) (string, tgbotapi.InlineKeyboardMarkup) {
	// Получаем текущую конфигурацию
	config := GetTriggerConfig()
	if config == nil || len(config) == 0 {
		return "❌ Триггеры не загружены", tgbotapi.NewInlineKeyboardMarkup()
	}

	totalTriggers := len(config)
	startIdx := page * triggersPerPage
	endIdx := startIdx + triggersPerPage
	
	// Проверяем границы
	if startIdx >= totalTriggers {
		startIdx = 0
		page = 0
		endIdx = triggersPerPage
	}
	if endIdx > totalTriggers {
		endIdx = totalTriggers
	}

	// Формируем заголовок
	header := fmt.Sprintf("📋 Триггеры %d-%d из %d:\n\n", 
		startIdx+1, endIdx, totalTriggers)

	// Создаем кнопки для триггеров текущей страницы
	var buttonRows [][]tgbotapi.InlineKeyboardButton
	
	for i := startIdx; i < endIdx; i++ {
		trigger := config[i]
		triggerNum := i + 1
		
		// Форматируем текст кнопки
		buttonText := formatTriggerButton(trigger, triggerNum)
		
		// Создаем callback_data
		callbackData := fmt.Sprintf("trigger_info:%s", trigger.TechKey)
		
		// Создаем кнопку (одна кнопка в ряд)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Добавляем навигацию (если нужно)
	navRow := createNavigationButtons(page, totalTriggers)
	if len(navRow) > 0 {
		buttonRows = append(buttonRows, navRow)
	}

	return header, tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// formatTriggerButton форматирует текст для кнопки триггера
func formatTriggerButton(trigger Trigger, number int) string {
	// Обрезаем название если нужно
	displayName := trigger.TriggerName
	if len(displayName) > maxNameLength {
		displayName = displayName[:maxNameLength-3] + "..."
	}
	
	// Статистика триггера
	patternsCount := len(trigger.Patterns)
	responsesCount := len(trigger.Responses)
	probability := int(trigger.Probability * 100)
	
	return fmt.Sprintf("%d. %s (%d%%, %d, %d)", 
		number, displayName, probability, patternsCount, responsesCount)
}

// createNavigationButtons создает кнопки навигации
func createNavigationButtons(currentPage, totalTriggers int) []tgbotapi.InlineKeyboardButton {
	var buttons []tgbotapi.InlineKeyboardButton
	
	// Проверяем, есть ли следующая страница
	totalPages := (totalTriggers + triggersPerPage - 1) / triggersPerPage
	hasNextPage := (currentPage+1) < totalPages
	
	if hasNextPage {
		callbackData := fmt.Sprintf("triggers_page:%d", currentPage+1)
		button := tgbotapi.NewInlineKeyboardButtonData("⏩ Далее", callbackData)
		buttons = append(buttons, button)
	}
	
	return buttons
}
