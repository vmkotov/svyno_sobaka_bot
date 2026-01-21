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

// GenerateTriggersMenu создает меню с триггерами для указанной страницы
// Возвращает текст сообщения и inline-клавиатуру
func GenerateTriggersMenu(page int) (string, tgbotapi.InlineKeyboardMarkup) {
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

		// Создаем callback_data по новой системе
		callbackData := fmt.Sprintf("trigger:detail:%s", trigger.TechKey)

		// Создаем кнопку (одна кнопка в ряд)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Добавляем навигацию (всегда показываем "Главная")
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

	// Рассчитываем общее количество страниц
	totalPages := (totalTriggers + triggersPerPage - 1) / triggersPerPage

	// Определяем, какие кнопки показывать
	hasPrevPage := currentPage > 0
	hasNextPage := (currentPage + 1) < totalPages

	// Кнопка "Назад" (если не первая страница)
	if hasPrevPage {
		callbackData := fmt.Sprintf("triggers:page:%d", currentPage-1)
		button := tgbotapi.NewInlineKeyboardButtonData("⏪ Назад", callbackData)
		buttons = append(buttons, button)
	}

	// Кнопка "Главная" (ВСЕГДА показываем!)
	callbackData := "menu:main"
	button := tgbotapi.NewInlineKeyboardButtonData("🏠 Главная", callbackData)
	buttons = append(buttons, button)

	// Кнопка "Далее" (если не последняя страница)
	if hasNextPage {
		callbackData := fmt.Sprintf("triggers:page:%d", currentPage+1)
		button := tgbotapi.NewInlineKeyboardButtonData("⏩ Далее", callbackData)
		buttons = append(buttons, button)
	}

	return buttons
}

// GenerateAdminTriggersMenu создает админское меню с триггерами для указанной страницы
// Возвращает текст сообщения и inline-клавиатуру с админской навигацией
func GenerateAdminTriggersMenu(page int) (string, tgbotapi.InlineKeyboardMarkup) {
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
	header := fmt.Sprintf("📋 *Админка - Триггеры %d-%d из %d:*\n\n",
		startIdx+1, endIdx, totalTriggers)

	// Создаем кнопки для триггеров текущей страницы
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	for i := startIdx; i < endIdx; i++ {
		trigger := config[i]
		triggerNum := i + 1

		// Форматируем текст кнопки
		buttonText := formatTriggerButton(trigger, triggerNum)

		// Создаем callback_data по админской системе
		callbackData := fmt.Sprintf("admin:trigger:detail:%s", trigger.TechKey)

		// Создаем кнопку (одна кнопка в ряд)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Добавляем админскую навигацию
	navRow := createAdminNavigationButtons(page, totalTriggers)
	if len(navRow) > 0 {
		buttonRows = append(buttonRows, navRow)
	}

	return header, tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// createAdminNavigationButtons создает кнопки навигации для админки
func createAdminNavigationButtons(currentPage, totalTriggers int) []tgbotapi.InlineKeyboardButton {
	var buttons []tgbotapi.InlineKeyboardButton

	// Рассчитываем общее количество страниц
	totalPages := (totalTriggers + triggersPerPage - 1) / triggersPerPage

	// Определяем, какие кнопки показывать
	hasPrevPage := currentPage > 0
	hasNextPage := (currentPage + 1) < totalPages

	// Кнопка "Назад" (если не первая страница)
	if hasPrevPage {
		callbackData := fmt.Sprintf("admin:triggers:page:%d", currentPage-1)
		button := tgbotapi.NewInlineKeyboardButtonData("⏪ Назад", callbackData)
		buttons = append(buttons, button)
	}

	// Кнопка "В админку" (ВСЕГДА показываем!)
	callbackData := "admin:menu"
	button := tgbotapi.NewInlineKeyboardButtonData("🐷 В админку", callbackData)
	buttons = append(buttons, button)

	// Кнопка "Далее" (если не последняя страница)
	if hasNextPage {
		callbackData := fmt.Sprintf("admin:triggers:page:%d", currentPage+1)
		button := tgbotapi.NewInlineKeyboardButtonData("⏩ Далее", callbackData)
		buttons = append(buttons, button)
	}

	return buttons
}
