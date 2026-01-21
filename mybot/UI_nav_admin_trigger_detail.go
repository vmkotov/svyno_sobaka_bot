package mybot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GenerateTriggerDetailCard создает детальную карточку триггера
func GenerateTriggerDetailCard(trigger *Trigger, fromPage int) (string, tgbotapi.InlineKeyboardMarkup) {
	if trigger == nil {
		return createErrorMessage("unknown"), createBackButton(fromPage)
	}

	// Форматируем детали
	message := formatTriggerDetail(trigger)

	// Логируем сообщение для отладки
	log.Printf("🔍 Детальная карточка для %s, длина: %d байт",
		trigger.TriggerName, len(message))

	// Проверим Markdown проблемы
	if strings.Count(message, "*")%2 != 0 {
		log.Printf("⚠️ Нечетное количество звёздочек в Markdown: %d",
			strings.Count(message, "*"))
	}

	keyboard := createDetailKeyboard(trigger.TechKey, fromPage)

	return message, keyboard
}

// GenerateAdminTriggerDetailCard создает админскую детальную карточку триггера
func GenerateAdminTriggerDetailCard(trigger *Trigger, fromPage int) (string, tgbotapi.InlineKeyboardMarkup) {
	if trigger == nil {
		return createErrorMessage("unknown"), createAdminBackButton(fromPage)
	}

	// Форматируем детали
	message := formatTriggerDetail(trigger)

	// Добавляем админскую пометку
	message = "👑 *АДМИНКА*\n\n" + message

	// Логируем сообщение для отладки
	log.Printf("👑 Админская детальная карточка для %s, длина: %d байт",
		trigger.TriggerName, len(message))

	keyboard := createAdminDetailKeyboard(trigger.TechKey, fromPage)

	return message, keyboard
}

func createErrorMessage(techKey string) string {
	return fmt.Sprintf("❌ Триггер с ключом `%s` не найден\n\n"+
		"Возможно, он был удален или изменен. "+
		"Используйте /refresh_me чтобы обновить список.", safeCode(techKey))
}

func formatTriggerDetail(trigger *Trigger) string {
	// Форматируем паттерны с умным экранированием
	patternsText := formatPatterns(trigger.Patterns)

	// Форматируем ответы с умным экранированием
	responsesText := formatResponses(trigger.Responses)

	// Основное сообщение - используем safeMarkdown для текста
	return fmt.Sprintf(
		"🎯 *%s*\n\n"+
			"🔑 Тех. ключ: `%s`\n"+
			"🎯 Приоритет: %d\n"+
			"🎲 Вероятность: %d%%\n"+
			"📊 Паттернов: %d | Ответов: %d\n\n"+
			"🔍 *Паттерны:*\n%s\n\n"+
			"💬 *Ответы:*\n%s\n\n"+
			"Ключ: `%s`",
		safeMarkdown(trigger.TriggerName),
		safeCode(trigger.TechKey),
		trigger.Priority,
		int(trigger.Probability*100),
		len(trigger.Patterns),
		len(trigger.Responses),
		patternsText,
		responsesText,
		safeCode(trigger.TechKey),
	)
}

func formatPatterns(patterns []Pattern) string {
	if len(patterns) == 0 {
		return "Нет паттернов"
	}

	var builder strings.Builder
	for i, p := range patterns {
		escapedPattern := safeCode(p.PatternText)
		builder.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, escapedPattern))
	}
	return builder.String()
}

func formatResponses(responses []Response) string {
	if len(responses) == 0 {
		return "Нет ответов"
	}

	var builder strings.Builder
	for i, r := range responses {
		escapedResponse := safeMarkdown(r.ResponseText)
		builder.WriteString(fmt.Sprintf("%d. %s (вес: %d)\n",
			i+1, escapedResponse, r.ResponseWeight))
	}
	return builder.String()
}

func createDetailKeyboard(techKey string, fromPage int) tgbotapi.InlineKeyboardMarkup {
	backCallback := fmt.Sprintf("triggers:page:%d", fromPage)
	homeCallback := "menu:main"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", backCallback),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главная", homeCallback),
		),
	)

	return keyboard
}

func createBackButton(fromPage int) tgbotapi.InlineKeyboardMarkup {
	backCallback := fmt.Sprintf("triggers:page:%d", fromPage)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", backCallback),
		),
	)

	return keyboard
}

func createAdminDetailKeyboard(techKey string, fromPage int) tgbotapi.InlineKeyboardMarkup {
	backCallback := fmt.Sprintf("admin:triggers:page:%d", fromPage)
	adminCallback := "admin:menu"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", backCallback),
			tgbotapi.NewInlineKeyboardButtonData("🐷 В админку", adminCallback),
		),
	)

	return keyboard
}

func createAdminBackButton(fromPage int) tgbotapi.InlineKeyboardMarkup {
	backCallback := fmt.Sprintf("admin:triggers:page:%d", fromPage)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", backCallback),
		),
	)

	return keyboard
}

func extractPageFromMessage(text string) int {
	return 0
}
