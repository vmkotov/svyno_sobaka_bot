package mybot

import (
	"fmt"
	"strings"
)

// getEmojiNumber возвращает эмодзи-цифру для номера (динамически)
func getEmojiNumber(n int) string {
	if n < 1 {
		return fmt.Sprintf("%d.", n)
	}

	// Мапим цифры на эмодзи
	digitEmoji := map[rune]string{
		'0': "0️⃣",
		'1': "1️⃣",
		'2': "2️⃣",
		'3': "3️⃣",
		'4': "4️⃣",
		'5': "5️⃣",
		'6': "6️⃣",
		'7': "7️⃣",
		'8': "8️⃣",
		'9': "9️⃣",
	}

	// Преобразуем число в строку и заменяем каждую цифру эмодзи
	numStr := fmt.Sprintf("%d", n)
	var result strings.Builder

	for _, digit := range numStr {
		if emoji, ok := digitEmoji[digit]; ok {
			result.WriteString(emoji)
		} else {
			result.WriteRune(digit)
		}
	}

	return result.String()
}

// FormatTriggerStats форматирует статистику триггеров
func FormatTriggerStats(config []Trigger) string {
	totalPatterns := 0
	totalResponses := 0

	for _, trigger := range config {
		totalPatterns += len(trigger.Patterns)
		totalResponses += len(trigger.Responses)
	}

	return fmt.Sprintf("📊 Статистика:\n"+
		"• Всего триггеров: %d\n"+
		"• Всего паттернов: %d\n"+
		"• Всего ответов: %d",
		len(config), totalPatterns, totalResponses)
}

// FormatTriggersList форматирует список триггеров
func FormatTriggersList(config []Trigger) string {
	var builder strings.Builder

	builder.WriteString("📋 Список по приоритету:\n")

	for i, trigger := range config {
		// Формат: 1️⃣ Название (100%, 2, 2)
		builder.WriteString(fmt.Sprintf("%s %s (%d%%, %d, %d)\n",
			getEmojiNumber(i+1),
			trigger.TriggerName,
			int(trigger.Probability*100),
			len(trigger.Patterns),
			len(trigger.Responses)))
	}

	return builder.String()
}

// SplitLongMessage разбивает длинное сообщение на части по maxLen
func SplitLongMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var parts []string
	lines := strings.Split(text, "\n")
	var currentPart strings.Builder

	for _, line := range lines {
		if currentPart.Len()+len(line)+1 > maxLen && currentPart.Len() > 0 {
			parts = append(parts, currentPart.String())
			currentPart.Reset()
		}

		if currentPart.Len() > 0 {
			currentPart.WriteByte('\n')
		}
		currentPart.WriteString(line)
	}

	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	return parts
}
