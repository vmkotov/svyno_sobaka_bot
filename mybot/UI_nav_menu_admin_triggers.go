package mybot

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для меню триггеров
const (
	triggersPerPage = 10 // Триггеров на страницу
	maxNameLength   = 25 // Максимальная длина названия в кнопке
)

// GenerateTriggersMenu создает пользовательское меню триггеров
func GenerateTriggersMenu(page int) (string, tgbotapi.InlineKeyboardMarkup) {
	// Получаем текущую конфигурация
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
		callbackData := fmt.Sprintf("trigger:detail:%s", trigger.TechKey)

		// Создаем кнопку (одна кнопка в ряд)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Добавляем навигацию
	navRow := createNavigationButtons(page, totalTriggers)
	if len(navRow) > 0 {
		buttonRows = append(buttonRows, navRow)
	}

	return header, tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
}

// GenerateAdminTriggersMenu создает админское меню с триггерами для указанной страницы
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

	// ДОБАВЛЯЕМ КНОПКУ СОЗДАНИЯ НОВОГО ТРИГГЕРА
	newTriggerBtn := tgbotapi.NewInlineKeyboardButtonData("➕ Добавить сценарий", "admin:trigger:new")
	buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(newTriggerBtn))

	// Добавляем админскую навигацию
	navRow := createAdminNavigationButtons(page, totalTriggers)
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

// HandleAdminTriggerDetailCallback - обработка admin:trigger:detail:TECH_KEY
func HandleAdminTriggerDetailCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	if len(parts) < 4 {
		log.Printf("⚠️ Неполный callback_data для деталей триггера: %v", parts)
		return
	}

	// Получаем триггер
	techKey := parts[3]
	trigger := GetTriggerByTechKey(techKey)

	if trigger == nil {
		log.Printf("❌ Триггер с ключом %s не найден", techKey)
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Триггер не найден")
		bot.Request(callback)
		return
	}

	log.Printf("👑 Админская детальная карточка триггера %s от @%s",
		techKey, callbackQuery.From.UserName)

	// Извлекаем номер страницы
	fromPage := extractPageFromMessage(callbackQuery.Message.Text)

	// Генерируем админскую детальную карточку
	message, keyboard := GenerateAdminTriggerDetailCard(trigger, fromPage)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		message,
		keyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки админской детальной карточки: %v", err)
	}
}

// extractPageFromMessage извлекает номер страницы из текста сообщения
