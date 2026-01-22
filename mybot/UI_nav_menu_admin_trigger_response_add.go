// ============================================================================
// ФАЙЛ: UI_nav_menu_admin_trigger_response_add.go
// Обработка добавления ответа к триггеру
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

// Структура для состояния добавления ответа
type ResponseAddState struct {
	TechKey   string    // Технический ключ триггера
	UserID    int64     // ID пользователя
	ChatID    int64     // ID чата
	MessageID int64     // ID сообщения с формой
	CreatedAt time.Time // Время создания формы
}

// Карта состояний (временное решение)
var responseAddStates = make(map[int64]*ResponseAddState) // key: userID

// handleAddResponse - показывает форму добавления ответа
func handleAddResponse(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, techKey string) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	bot.Request(callback)

	log.Printf("🛠️ Показать форму добавления ответа для %s от @%s",
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
		"✏️ *Добавление ответа*\n\n"+
			"Триггер: *%s*\n"+
			"Ключ: `%s`\n\n"+
			"Введите новый ответ:\n"+
			"_Например: \"Сам такой!\", \"Это точно!\"_\n\n"+
			"⚠️ Ответ должен быть от 2 до 100 символов",
		safeMarkdown(trigger.TriggerName),
		safeCode(techKey),
	)

	// Создаем inline-клавиатуру
	cancelCallback := fmt.Sprintf("admin:trigger:response:cancel:%s", techKey)
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
	state := &ResponseAddState{
		TechKey:   techKey,
		UserID:    callbackQuery.From.ID,
		ChatID:    callbackQuery.Message.Chat.ID,
		MessageID: int64(sentMsg.MessageID),
		CreatedAt: time.Now(),
	}
	responseAddStates[callbackQuery.From.ID] = state

	log.Printf("✅ Форма добавления ответа отправлена для @%s (message_id: %d)",
		callbackQuery.From.UserName, sentMsg.MessageID)
}

// handleAddResponseCancel - отмена добавления ответа
func handleAddResponseCancel(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, techKey string) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Добавление отменено")
	bot.Request(callback)

	log.Printf("❌ Отмена добавления ответа для %s от @%s",
		techKey, callbackQuery.From.UserName)

	// Удаляем форму сообщения
	msg := tgbotapi.NewDeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
	bot.Send(msg)

	// Очищаем состояние
	delete(responseAddStates, callbackQuery.From.ID)

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

// ProcessResponseInput - обработка ввода ответа пользователем
func ProcessResponseInput(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) bool {
	// Проверяем, есть ли состояние добавления ответа для этого пользователя
	state, exists := responseAddStates[msg.From.ID]
	if !exists {
		return false // Это не ввод ответа
	}

	// Проверяем что сообщение в том же чате
	if msg.Chat.ID != state.ChatID {
		log.Printf("⚠️ Сообщение не из того чата для состояния ответа")
		return false
	}

	log.Printf("📝 Обработка ввода ответа от @%s: %s",
		msg.From.UserName, msg.Text)

	// Валидация ответа
	responseText := strings.TrimSpace(msg.Text)
	if len(responseText) < 2 {
		SendMessage(bot, msg.Chat.ID, "❌ Ответ должен быть не менее 2 символов", "ошибка валидации")
		return true
	}
	if len(responseText) > 100 {
		SendMessage(bot, msg.Chat.ID, "❌ Ответ должен быть не более 100 символов", "ошибка валидации")
		return true
	}

	// Удаляем сообщение пользователя (чтобы не засорять чат)
	deleteMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)
	bot.Send(deleteMsg)

	// Создаем скрытое сообщение с ответом
	hiddenMsg := tgbotapi.NewMessage(msg.Chat.ID, responseText)
	hiddenMsg.DisableNotification = true
	sentHiddenMsg, err := bot.Send(hiddenMsg)
	if err != nil {
		log.Printf("❌ Ошибка создания скрытого сообщения: %v", err)
		SendMessage(bot, msg.Chat.ID, "❌ Ошибка при обработке ответа", "ошибка")
		delete(responseAddStates, msg.From.ID)
		return true
	}

	// Вызываем процедуру БД
	log.Printf("📊 Вызов процедуры для триггера %s с ответом: %s (message_id: %d)",
		state.TechKey, responseText, sentHiddenMsg.MessageID)

	// Проверяем что БД доступна
	if db == nil {
		log.Printf("❌ БД не доступна для сохранения ответа")

		// Удаляем скрытое сообщение
		deleteHiddenMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentHiddenMsg.MessageID)
		bot.Send(deleteHiddenMsg)

		showResponseAddResult(bot, state, responseText, false, "БД не доступна")
		delete(responseAddStates, msg.From.ID)
		return true
	}

	// РЕАЛЬНЫЙ ВЫЗОВ ПРОЦЕДУРЫ
	// Третий параметр - это message_id для логов, но для ответов передаем nil
	_, err = db.Exec("CALL svyno_sobaka_bot.proc_insert_response($1, $2, $3)",
		state.TechKey, responseText, nil)

	if err != nil {
		log.Printf("❌ Ошибка вызова процедуры БД: %v", err)

		// Удаляем скрытое сообщение при ошибке
		deleteHiddenMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentHiddenMsg.MessageID)
		bot.Send(deleteHiddenMsg)

		showResponseAddResult(bot, state, responseText, false, "Ошибка БД: "+err.Error())
		delete(responseAddStates, msg.From.ID)
		return true
	}

	log.Printf("✅ Процедура успешно вызвана для триггера %s", state.TechKey)

	// Удаляем скрытое сообщение
	deleteHiddenMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentHiddenMsg.MessageID)
	bot.Send(deleteHiddenMsg)

	// Обновляем форму с результатом
	showResponseAddResult(bot, state, responseText, true, "")

	// Очищаем состояние
	delete(responseAddStates, msg.From.ID)

	return true
}

// showResponseAddResult - показывает результат добавления ответа
func showResponseAddResult(bot *tgbotapi.BotAPI, state *ResponseAddState,
	responseText string, success bool, errorMsg string) {

	trigger := GetTriggerByTechKey(state.TechKey)
	if trigger == nil {
		log.Printf("❌ Триггер %s не найден для показа результата", state.TechKey)
		return
	}

	var resultText string
	if success {
		resultText = fmt.Sprintf(
			"✅ *Ответ добавлен!*\n\n"+
				"Триггер: *%s*\n"+
				"Ответ: `%s`\n\n"+
				"Теперь триггер имеет %d ответов",
			safeMarkdown(trigger.TriggerName),
			safeCode(responseText),
			len(trigger.Responses)+1, // +1 новый ответ
		)
	} else {
		resultText = fmt.Sprintf(
			"❌ *Ошибка добавления ответа*\n\n"+
				"Триггер: *%s*\n"+
				"Ответ: `%s`\n\n"+
				"Ошибка: %s",
			safeMarkdown(trigger.TriggerName),
			safeCode(responseText),
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

// cleanupResponseStates - очистка устаревших состояний
func cleanupResponseStates() {
	now := time.Now()
	for userID, state := range responseAddStates {
		if now.Sub(state.CreatedAt) > 5*time.Minute {
			log.Printf("🧹 Очистка устаревшего состояния ответа для user_id: %d", userID)
			delete(responseAddStates, userID)
		}
	}
}
