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
	keyboard := createDetailKeyboard(trigger.TechKey, fromPage)

	return message, keyboard
}

// HandleTriggerDetailCallback обрабатывает callback детальной карточки
func HandleTriggerDetailCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	bot.Request(callback)

	if len(parts) < 3 {
		log.Printf("⚠️ Неполный callback_data для деталей триггера: %v", parts)
		return
	}

	techKey := parts[2] // format: "trigger:detail:tech_key"

	// Извлекаем номер страницы из сообщения или используем 0
	fromPage := extractPageFromMessage(callbackQuery.Message.Text)

	// Получаем триггер
	trigger := GetTriggerByTechKey(techKey)

	// Генерируем детальную карточку
	message, keyboard := GenerateTriggerDetailCard(trigger, fromPage)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		message,
		keyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки детальной карточки: %v", err)
	}
}

// ================= ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ =================

func createErrorMessage(techKey string) string {
	return fmt.Sprintf("❌ Триггер с ключом `%s` не найден\n\n"+
		"Возможно, он был удален или изменен. "+
		"Используйте /refresh_me чтобы обновить список.", techKey)
}

func formatTriggerDetail(trigger *Trigger) string {
	// Форматируем паттерны
	patternsText := formatPatterns(trigger.Patterns)

	// Форматируем ответы
	responsesText := formatResponses(trigger.Responses)

	// Основное сообщение
	return fmt.Sprintf(
		"🎯 *%s*\n\n"+
			"🔑 Тех. ключ: `%s`\n"+
			"🎯 Приоритет: %d\n"+
			"🎲 Вероятность: %d%%\n"+
			"📊 Паттернов: %d | Ответов: %d\n\n"+
			"🔍 *Паттерны:*\n%s\n\n"+
			"💬 *Ответы:*\n%s\n\n"+
			"#%s",
		trigger.TriggerName,
		trigger.TechKey,
		trigger.Priority,
		int(trigger.Probability*100),
		len(trigger.Patterns),
		len(trigger.Responses),
		patternsText,
		responsesText,
		trigger.TechKey,
	)
}

func formatPatterns(patterns []Pattern) string {
	if len(patterns) == 0 {
		return "Нет паттернов"
	}

	var builder strings.Builder
	for i, p := range patterns {
		builder.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, p.PatternText))
	}
	return builder.String()
}

func formatResponses(responses []Response) string {
	if len(responses) == 0 {
		return "Нет ответов"
	}

	var builder strings.Builder
	for i, r := range responses {
		builder.WriteString(fmt.Sprintf("%d. %s (вес: %d)\n",
			i+1, r.ResponseText, r.ResponseWeight))
	}
	return builder.String()
}

func createDetailKeyboard(techKey string, fromPage int) tgbotapi.InlineKeyboardMarkup {
	// Кнопка "Назад" возвращает на ту же страницу
	backCallback := fmt.Sprintf("triggers:page:%d", fromPage)

	// Кнопка "Главная"
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

// extractPageFromMessage пытается извлечь номер страницы из текста сообщения
func extractPageFromMessage(text string) int {
	// Простая реализация - всегда возвращаем 0
	// TODO: можно добавить парсинг "Триггеры 1-10 из 50"
	return 0
}
