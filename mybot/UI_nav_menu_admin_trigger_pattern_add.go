// ============================================================================
// ФАЙЛ: UI_nav_menu_admin_trigger_pattern_add.go
// Обработка добавления паттерна к триггеру
// ============================================================================
package mybot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структура для состояния добавления паттерна
type PatternAddState struct {
	TechKey   string    // Технический ключ триггера
	UserID    int64     // ID пользователя
	ChatID    int64     // ID чата
	MessageID int64     // ID сообщения с формой
	CreatedAt time.Time // Время создания формы
}

// Карта состояний (временное решение)
var patternAddStates = make(map[int64]*PatternAddState) // key: userID

// handleAddPattern - показывает форму добавления паттерна
func handleAddPattern(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, techKey string) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	bot.Request(callback)

	log.Printf("🛠️ Показать форму добавления паттерна для %s от @%s",
		techKey, callbackQuery.From.UserName)

	// Получаем триггер для отображения названия
	trigger := GetTriggerByTechKey(techKey)
	if trigger == nil {
		log.Printf("❌ Триггер с ключом %s не найден", techKey)
		callback = tgbotapi.NewCallback(callbackQuery.ID, "❌ Триггер не найден")
		bot.Request(callback)
		return
	}

	// Создаем сообщение с формой
	formText := fmt.Sprintf(
		"✏️ *Добавление паттерна*\n\n"+
			"Триггер: *%s*\n"+
			"Ключ: `%s`\n\n"+
			"Введите новый паттерн:\n"+
			"_Например: \"прикол\", \"смешно\", \"ржач\"_\n\n"+
			"⚠️ Паттерн должен быть от 2 до 100 символов",
		safeMarkdown(trigger.TriggerName),
		safeCode(techKey),
	)

	// Создаем inline-клавиатуру
	cancelCallback := fmt.Sprintf("admin:trigger:pattern:cancel:%s", techKey)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", cancelCallback),
		),
	)

	// Отправляем форму как новое сообщение
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, formText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Printf("❌ Ошибка отправки формы: %v", err)
		return
	}

	// Сохраняем состояние
	state := &PatternAddState{
		TechKey:   techKey,
		UserID:    callbackQuery.From.ID,
		ChatID:    callbackQuery.Message.Chat.ID,
		MessageID: int64(sentMsg.MessageID),
		CreatedAt: time.Now(),
	}
	patternAddStates[callbackQuery.From.ID] = state

	log.Printf("✅ Форма добавления паттерна отправлена для @%s (message_id: %d)",
		callbackQuery.From.UserName, sentMsg.MessageID)
}

// handleAddPatternCancel - отмена добавления паттерна
func handleAddPatternCancel(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, techKey string) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Добавление отменено")
	bot.Request(callback)

	log.Printf("❌ Отмена добавления паттерна для %s от @%s",
		techKey, callbackQuery.From.UserName)

	// Удаляем форму сообщения
	msg := tgbotapi.NewDeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
	bot.Send(msg)

	// Очищаем состояние
	delete(patternAddStates, callbackQuery.From.ID)

	// Возвращаем в детальную карточку триггера
	trigger := GetTriggerByTechKey(techKey)
	if trigger != nil {
		message, keyboard := GenerateAdminTriggerDetailCard(trigger, 0)

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			message,
			keyboard,
		)
		editMsg.ParseMode = "Markdown"
		bot.Send(editMsg)
	}
}

// ProcessPatternInput - обработка ввода паттерна пользователем
func ProcessPatternInput(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) bool {
	// Проверяем, есть ли состояние добавления паттерна для этого пользователя
	state, exists := patternAddStates[msg.From.ID]
	if !exists {
		return false // Это не ввод паттерна
	}

	// Проверяем что сообщение в том же чате
	if msg.Chat.ID != state.ChatID {
		log.Printf("⚠️ Сообщение не из того чата для состояния паттерна")
		return false
	}

	log.Printf("📝 Обработка ввода паттерна от @%s: %s",
		msg.From.UserName, msg.Text)

	// Валидация паттерна
	patternText := strings.TrimSpace(msg.Text)
	if len(patternText) < 2 {
		SendMessage(bot, msg.Chat.ID, "❌ Паттерн должен быть не менее 2 символов", "ошибка валидации")
		return true
	}
	if len(patternText) > 100 {
		SendMessage(bot, msg.Chat.ID, "❌ Паттерн должен быть не более 100 символов", "ошибка валидации")
		return true
	}

	// Удаляем сообщение пользователя (чтобы не засорять чат)
	deleteMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)
	bot.Send(deleteMsg)

	// Создаем скрытое сообщение с паттерном
	hiddenMsg := tgbotapi.NewMessage(msg.Chat.ID, patternText)
	hiddenMsg.DisableNotification = true
	sentHiddenMsg, err := bot.Send(hiddenMsg)
	if err != nil {
		log.Printf("❌ Ошибка создания скрытого сообщения: %v", err)
		SendMessage(bot, msg.Chat.ID, "❌ Ошибка при обработке паттерна", "ошибка")
		delete(patternAddStates, msg.From.ID)
		return true
	}

	// Вызываем процедуру БД
	log.Printf("📊 Вызов процедуры для триггера %s с паттерном: %s (message_id: %d)",
		state.TechKey, patternText, sentHiddenMsg.MessageID)

	// Проверяем что БД доступна
	if db == nil {
		log.Printf("❌ БД не доступна для сохранения паттерна")

		// Удаляем скрытое сообщение
		deleteHiddenMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentHiddenMsg.MessageID)
		bot.Send(deleteHiddenMsg)

		showPatternAddResult(bot, state, patternText, false, "БД не доступна")
		delete(patternAddStates, msg.From.ID)
		return true
	}

	// РЕАЛЬНЫЙ ВЫЗОВ ПРОЦЕДУРЫ
	_, err = db.Exec("CALL svyno_sobaka_bot.update_pattern_with_logging($1, $2, $3)",
		state.TechKey, patternText, sentHiddenMsg.MessageID)

	if err != nil {
		log.Printf("❌ Ошибка вызова процедуры БД: %v", err)

		// Удаляем скрытое сообщение при ошибке
		deleteHiddenMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentHiddenMsg.MessageID)
		bot.Send(deleteHiddenMsg)

		showPatternAddResult(bot, state, patternText, false, "Ошибка БД: "+err.Error())
		delete(patternAddStates, msg.From.ID)
		return true
	}

	log.Printf("✅ Процедура успешно вызвана для триггера %s", state.TechKey)

	// Удаляем скрытое сообщение
	deleteHiddenMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentHiddenMsg.MessageID)
	bot.Send(deleteHiddenMsg)

	// Обновляем форму с результатом
	showPatternAddResult(bot, state, patternText, true, "")

	// Очищаем состояние
	delete(patternAddStates, msg.From.ID)

	return true
}

// showPatternAddResult - показывает результат добавления паттерна
func showPatternAddResult(bot *tgbotapi.BotAPI, state *PatternAddState,
	patternText string, success bool, errorMsg string) {

	trigger := GetTriggerByTechKey(state.TechKey)
	if trigger == nil {
		log.Printf("❌ Триггер %s не найден для показа результата", state.TechKey)
		return
	}

	var resultText string
	if success {
		resultText = fmt.Sprintf(
			"✅ *Паттерн добавлен!*\n\n"+
				"Триггер: *%s*\n"+
				"Паттерн: `%s`\n\n"+
				"Теперь триггер ищет %d паттернов",
			safeMarkdown(trigger.TriggerName),
			safeCode(patternText),
			len(trigger.Patterns)+1, // +1 новый паттерн
		)
	} else {
		resultText = fmt.Sprintf(
			"❌ *Ошибка добавления паттерна*\n\n"+
				"Триггер: *%s*\n"+
				"Паттерн: `%s`\n\n"+
				"Ошибка: %s",
			safeMarkdown(trigger.TriggerName),
			safeCode(patternText),
			errorMsg,
		)
	}

	// Кнопки после добавления
	refreshCallback := fmt.Sprintf("admin:trigger:detail:%s", state.TechKey)
	adminCallback := "admin:menu"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить карточку", refreshCallback),
			tgbotapi.NewInlineKeyboardButtonData("🏠 В админку", adminCallback),
		),
	)

	// Обновляем сообщение с формой
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		state.ChatID,
		int(state.MessageID),
		resultText,
		keyboard,
	)
	editMsg.ParseMode = "Markdown"

	if _, err := bot.Send(editMsg); err != nil {
		log.Printf("❌ Ошибка обновления формы результата: %v", err)
	}
}

// cleanupPatternStates - очистка устаревших состояний
func cleanupPatternStates() {
	now := time.Now()
	for userID, state := range patternAddStates {
		if now.Sub(state.CreatedAt) > 5*time.Minute {
			log.Printf("🧹 Очистка устаревшего состояния паттерна для user_id: %d", userID)
			delete(patternAddStates, userID)
		}
	}
}
