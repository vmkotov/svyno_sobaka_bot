package mybot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CheckAllTriggers проверяет ВСЕ триггеры в порядке приоритета
// Возвращает true при первом срабатывании любого триггера
func CheckAllTriggers(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, logChatID int64, db *sql.DB) bool {
	if msg.Text == "" {
		return false
	}

	// ✅ Берем конфигурацию из памяти (уже загружена при старте)
	config := GetTriggerConfig()
	if config == nil || len(config) == 0 {
		log.Printf("⚠️ Конфигурация триггеров пуста или не загружена")
		return false
	}

	// Нормализуем текст (как в оригинальных модулях)
	text := normalizeText(msg.Text)

	// Проверяем триггеры в порядке приоритета (они уже отсортированы)
	for _, trigger := range config {
		if checkSingleTrigger(bot, msg, text, &trigger, logChatID) {
			return true // Триггер сработал, дальше не проверяем
		}
	}

	return false
}

// checkSingleTrigger проверяет один триггер
// Возвращает: true если паттерны найдены (даже если ответ не отправлен)
func checkSingleTrigger(bot *tgbotapi.BotAPI, msg *tgbotapi.Message,
	normalizedText string, trigger *Trigger, logChatID int64) bool {

	// 1. Проверяем все паттерны триггера
	foundPatterns := []string{}
	for _, pattern := range trigger.Patterns {
		if strings.Contains(normalizedText, strings.ToLower(pattern.PatternText)) {
			foundPatterns = append(foundPatterns, pattern.PatternText)
		}
	}

	// Если ни один паттерн не найден - пропскаем триггер
	if len(foundPatterns) == 0 {
		return false
	}

	log.Printf("🔍 Триггер %s (приоритет %d): найдено %d паттернов от @%s",
		trigger.TriggerName, trigger.Priority, len(foundPatterns), msg.From.UserName)

	// ТРИГГЕР СРАБОТАЛ! Возвращаем true в любом случае
	// Но сначала проверяем вероятность ответа

	// 2. Проверяем вероятность (если < 1.0)
	if trigger.Probability < 1.0 {
		if randSource.Float64() > trigger.Probability {
			log.Printf("🎲 Пропущен ОТВЕТ триггера %s (вероятность %.0f%%)",
				trigger.TriggerName, trigger.Probability*100)
			sendTriggerLogToChat(bot, msg, trigger, foundPatterns, false, -1, logChatID, "рандомайзер")
			return true // Триггер сработал, но ответ не отправлен
		}
	}

	// 3. Выбираем случайный ответ (если несколько)
	if len(trigger.Responses) == 0 {
		log.Printf("⚠️ У триггера %s нет ответов", trigger.TriggerName)
		sendTriggerLogToChat(bot, msg, trigger, foundPatterns, false, -1, logChatID, "нет ответов")
		return true // Триггер сработал, но нет ответов
	}

	responseIndex := selectWeightedResponse(trigger.Responses)
	response := trigger.Responses[responseIndex]

	// 4. Проверка длины сообщения (не более 88 символов для ответа)
	log.Printf("📏 Длина сообщения для триггера %s: %d символов (normalized: %d)",
		trigger.TriggerName, len([]rune(msg.Text)), len([]rune(normalizedText)))

	if len([]rune(msg.Text)) > 88 {
		log.Printf("📏 Пропущен ОТВЕТ триггера %s (длина сообщения %d > 88 символов)",
			trigger.TriggerName, len([]rune(msg.Text)))
		sendTriggerLogToChat(bot, msg, trigger, foundPatterns, false, responseIndex, logChatID, "длина > 88 символов")
		return true // Триггер сработал, но ответ не отправлен из-за длины
	}

	// 5. Отправляем ответ
	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, response.ResponseText)
	replyMsg.ReplyToMessageID = msg.MessageID

	// Проверяем, нужен ли Markdown (как в оригинальных триггерах)
	if strings.Contains(response.ResponseText, "*") ||
		strings.Contains(response.ResponseText, "_") ||
		strings.Contains(response.ResponseText, "`") {
		replyMsg.ParseMode = "Markdown"
	}

	if _, err := bot.Send(replyMsg); err != nil {
		log.Printf("❌ Ошибка отправки ответа триггера %s: %v",
			trigger.TriggerName, err)
		sendTriggerLogToChat(bot, msg, trigger, foundPatterns, false, responseIndex, logChatID, "ошибка отправки")
		return true // Триггер сработал, но ошибка отправки
	}

	log.Printf("✅ Отправлен ответ триггера %s: %.30s...",
		trigger.TriggerName, response.ResponseText)

	// 6. Логируем в лог-чат
	sendTriggerLogToChat(bot, msg, trigger, foundPatterns, true, responseIndex, logChatID, "")

	return true // Триггер сработал И ответ отправлен
}

// selectWeightedResponse выбирает ответ с учетом весов
func selectWeightedResponse(responses []Response) int {
	if len(responses) == 0 {
		return 0
	}

	if len(responses) == 1 {
		return 0
	}

	// Если все веса равны 0 или не указаны - равномерное распределение
	totalWeight := 0
	for _, resp := range responses {
		totalWeight += resp.ResponseWeight
	}

	if totalWeight == 0 {
		return randSource.Intn(len(responses))
	}

	// Взвешенный случайный выбор
	randomValue := randSource.Intn(totalWeight)
	currentWeight := 0

	for i, resp := range responses {
		currentWeight += resp.ResponseWeight
		if randomValue < currentWeight {
			return i
		}
	}

	return len(responses) - 1
}

// sendTriggerLogToChat логирует срабатывание триггера в отдельный чат
func sendTriggerLogToChat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message,
	trigger *Trigger, foundPatterns []string,
	responded bool, responseIndex int, logChatID int64, skipReason string) {

	var reactionStatus string
	if responded {
		reactionStatus = fmt.Sprintf("✅ *Отреагировал* (вероятность %.0f%%)",
			trigger.Probability*100)
	} else if skipReason != "" {
		// Показываем причину пропуска
		reactionStatus = fmt.Sprintf("⏸️ *Пропущено: %s*", skipReason)
	} else {
		reactionStatus = fmt.Sprintf("🎲 *Пропущено рандомайзером* (вероятность %.0f%%)",
			trigger.Probability*100)
	}

	// Обрезаем список паттернов если их много
	patternsForLog := foundPatterns
	if len(foundPatterns) > 5 {
		patternsForLog = foundPatterns[:5]
	}

	responseText := ""
	if responded && responseIndex >= 0 && responseIndex < len(trigger.Responses) {
		responseText = trigger.Responses[responseIndex].ResponseText
		if len(responseText) > 50 {
			responseText = responseText[:50] + "..."
		}
	} else if len(trigger.Responses) > 0 {
		responseText = trigger.Responses[0].ResponseText
		if len(responseText) > 50 {
			responseText = responseText[:50] + "..."
		}
	}

	// Формируем основную часть лога
	logText := fmt.Sprintf(
		"🔔 *Триггер: %s*\n\n"+
			"%s\n"+
			"📝 *Сообщение:* `%s`\n"+
			"👤 *Пользователь:* %s\n"+
			"💬 *Чат ID:* `%d`\n"+
			"🎯 *Найденные паттерны:* %v\n"+
			"📊 *Всего паттернов:* %d\n"+
			"💬 *Ответ:* %s",
		escapeMarkdownForLog(trigger.TriggerName),
		reactionStatus,
		escapeMarkdownForLog(msg.Text),
		escapeMarkdownForLog(msg.From.FirstName),
		msg.Chat.ID,
		patternsForLog,
		len(foundPatterns),
		escapeMarkdownForLog(responseText),
	)

	// Добавляем хеш-тег БЕЗ Markdown форматирования (просто текст)
	hashtag := "#" + trigger.TechKey
	logText += "\n\n" + hashtag

	logMsg := tgbotapi.NewMessage(logChatID, logText)
	logMsg.ParseMode = "Markdown"

	if _, err := bot.Send(logMsg); err != nil {
		log.Printf("❌ Ошибка отправки лога триггера: %v", err)
		// Попробуем отправить без Markdown
		logMsg.ParseMode = ""
		if _, err2 := bot.Send(logMsg); err2 != nil {
			log.Printf("❌ Ошибка даже без Markdown: %v", err2)
		}
	}
}

// escapeMarkdownForLog - безопасное экранирование для логов
func escapeMarkdownForLog(text string) string {
	// Отличается от обычного escapeMarkdown - не экранирует дефисы и точки
	if text == "" {
		return ""
	}

	// Минимальный набор символов для экранирования в логах
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "=", "|", "{", "}", "\\"}

	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}

	return result
}
