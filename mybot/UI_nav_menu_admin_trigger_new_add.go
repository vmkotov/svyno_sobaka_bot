// ============================================================================
// ФАЙЛ: UI_nav_menu_admin_trigger_new_add.go
// Обработка создания нового триггера (сценария)
// ============================================================================
package mybot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структура для состояния добавления нового триггера
type NewTriggerAddState struct {
	UserID    int64     // ID пользователя
	ChatID    int64     // ID чата
	MessageID int64     // ID сообщения с формой
	CreatedAt time.Time // Время создания формы
}

// Карта состояний (временное решение)
var newTriggerAddStates = make(map[int64]*NewTriggerAddState) // key: userID

// generateTechKey генерирует технический ключ из русского текста
func generateTechKey(russianText string) string {
	// Транслитерация кириллицы в латиницу
	translitMap := map[rune]string{
		'а': "A", 'б': "B", 'в': "V", 'г': "G", 'д': "D",
		'е': "E", 'ё': "YO", 'ж': "ZH", 'з': "Z", 'и': "I",
		'й': "Y", 'к': "K", 'л': "L", 'м': "M", 'н': "N",
		'о': "O", 'п': "P", 'р': "R", 'с': "S", 'т': "T",
		'у': "U", 'ф': "F", 'х': "KH", 'ц': "TS", 'ч': "CH",
		'ш': "SH", 'щ': "SCH", 'ъ': "", 'ы': "Y", 'ь': "",
		'э': "E", 'ю': "YU", 'я': "YA",
		// Заглавные буквы
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D",
		'Е': "E", 'Ё': "YO", 'Ж': "ZH", 'З': "Z", 'И': "I",
		'Й': "Y", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N",
		'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T",
		'У': "U", 'Ф': "F", 'Х': "KH", 'Ц': "TS", 'Ч': "CH",
		'Ш': "SH", 'Щ': "SCH", 'Ъ': "", 'Ы': "Y", 'Ь': "",
		'Э': "E", 'Ю': "YU", 'Я': "YA",
	}

	var result strings.Builder
	
	// Преобразуем каждый символ
	for _, char := range russianText {
		if replacement, ok := translitMap[char]; ok {
			result.WriteString(replacement)
		} else if unicode.IsLetter(char) && char <= 'z' {
			// Если это латинская буква - переводим в верхний регистр
			result.WriteString(strings.ToUpper(string(char)))
		} else if char == ' ' {
			// Пробелы заменяются на _
			result.WriteString("_")
		}
		// Все остальные символы (знаки препинания и т.д.) игнорируем
	}

	// Обрезаем до 40 символов
	techKey := result.String()
	if len(techKey) > 40 {
		techKey = techKey[:40]
	}
	
	// Если после преобразования пусто - генерируем случайный
	if techKey == "" {
		timestamp := time.Now().Unix()
		techKey = fmt.Sprintf("TRIGGER_%d", timestamp%10000)
	}
	
	return techKey
}

// handleAddNewTrigger - показывает форму создания нового триггера
func handleAddNewTrigger(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	bot.Request(callback)

	log.Printf("🆕 Показать форму создания нового триггера от @%s",
		callbackQuery.From.UserName)

	// Создаем сообщение с формой
	formText := "✏️ *Создание нового триггера*\n\n" +
		"Введите название триггера:\n" +
		"_Например: \"Триггер для мата\", \"Приветствия\", \"Смешные ответы\"_\n\n" +
		"⚠️ Название должно быть от 3 до 100 символов\n\n" +
		"Технический ключ будет сгенерирован автоматически."

	// Создаем inline-клавиатуру
	cancelCallback := "admin:trigger:new:cancel"
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
	state := &NewTriggerAddState{
		UserID:    callbackQuery.From.ID,
		ChatID:    callbackQuery.Message.Chat.ID,
		MessageID: int64(sentMsg.MessageID),
		CreatedAt: time.Now(),
	}
	newTriggerAddStates[callbackQuery.From.ID] = state

	log.Printf("✅ Форма создания триггера отправлена для @%s (message_id: %d)",
		callbackQuery.From.UserName, sentMsg.MessageID)
}

// handleAddNewTriggerCancel - отмена создания нового триггера
func handleAddNewTriggerCancel(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Создание отменено")
	bot.Request(callback)

	log.Printf("❌ Отмена создания триггера от @%s",
		callbackQuery.From.UserName)

	// Удаляем форму сообщения
	msg := tgbotapi.NewDeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
	bot.Send(msg)

	// Очищаем состояние
	delete(newTriggerAddStates, callbackQuery.From.ID)

	// Возвращаем в список триггеров
	showAdminTriggersMenu(bot, callbackQuery, nil, 0)
}

// ProcessNewTriggerInput - обработка ввода названия нового триггера
func ProcessNewTriggerInput(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) bool {
	// Проверяем, есть ли состояние создания триггера для этого пользователя
	state, exists := newTriggerAddStates[msg.From.ID]
	if !exists {
		return false // Это не ввод нового триггера
	}

	// Проверяем что сообщение в том же чате
	if msg.Chat.ID != state.ChatID {
		log.Printf("⚠️ Сообщение не из того чата для состояния нового триггера")
		return false
	}

	log.Printf("📝 Обработка ввода нового триггера от @%s: %s",
		msg.From.UserName, msg.Text)

	// Валидация названия
	triggerName := strings.TrimSpace(msg.Text)
	if len(triggerName) < 3 {
		SendMessage(bot, msg.Chat.ID, "❌ Название должно быть не менее 3 символов", "ошибка валидации")
		return true
	}
	if len(triggerName) > 100 {
		triggerName = triggerName[:100] // Обрезаем до 100 символов
	}

	// Генерируем tech_key
	techKey := generateTechKey(triggerName)
	log.Printf("🔑 Сгенерирован tech_key: %s для названия: %s", techKey, triggerName)

	// Удаляем сообщение пользователя (чтобы не засорять чат)
	deleteMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)
	bot.Send(deleteMsg)

	// Создаем скрытое сообщение с названием
	hiddenMsg := tgbotapi.NewMessage(msg.Chat.ID, triggerName)
	hiddenMsg.DisableNotification = true
	sentHiddenMsg, err := bot.Send(hiddenMsg)
	if err != nil {
		log.Printf("❌ Ошибка создания скрытого сообщения: %v", err)
		SendMessage(bot, msg.Chat.ID, "❌ Ошибка при обработке триггера", "ошибка")
		delete(newTriggerAddStates, msg.From.ID)
		return true
	}

	// Вызываем процедуру БД
	log.Printf("📊 Вызов процедуры для создания триггера: %s (tech_key: %s, message_id: %d)",
		triggerName, techKey, sentHiddenMsg.MessageID)

	// Проверяем что БД доступна
	if db == nil {
		log.Printf("❌ БД не доступна для создания триггера")

		// Удаляем скрытое сообщение
		deleteHiddenMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentHiddenMsg.MessageID)
		bot.Send(deleteHiddenMsg)

		showNewTriggerAddResult(bot, state, triggerName, techKey, false, "БД не доступна")
		delete(newTriggerAddStates, msg.From.ID)
		return true
	}

	// РЕАЛЬНЫЙ ВЫЗОВ ПРОЦЕДУРЫ
	_, err = db.Exec("CALL svyno_sobaka_bot.proc_insert_trigger($1, $2, $3)",
		techKey, triggerName, sentHiddenMsg.MessageID)

	if err != nil {
		log.Printf("❌ Ошибка вызова процедуры БД: %v", err)

		// Удаляем скрытое сообщение при ошибке
		deleteHiddenMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentHiddenMsg.MessageID)
		bot.Send(deleteHiddenMsg)

		showNewTriggerAddResult(bot, state, triggerName, techKey, false, "Ошибка БД: "+err.Error())
		delete(newTriggerAddStates, msg.From.ID)
		return true
	}

	log.Printf("✅ Триггер успешно создан: %s (tech_key: %s)", triggerName, techKey)

	// Удаляем скрытое сообщение
	deleteHiddenMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, sentHiddenMsg.MessageID)
	bot.Send(deleteHiddenMsg)

	// Обновляем конфигурацию триггеров
	if err := LoadTriggerConfig(db); err != nil {
		log.Printf("⚠️ Не удалось обновить конфигурацию триггеров: %v", err)
	}

	// Обновляем форму с результатом
	showNewTriggerAddResult(bot, state, triggerName, techKey, true, "")

	// Очищаем состояние
	delete(newTriggerAddStates, msg.From.ID)

	return true
}

// showNewTriggerAddResult - показывает результат создания триггера
func showNewTriggerAddResult(bot *tgbotapi.BotAPI, state *NewTriggerAddState,
	triggerName, techKey string, success bool, errorMsg string) {

	var resultText string
	if success {
		resultText = fmt.Sprintf(
			"✅ *Триггер создан!*\n\n"+
				"Название: *%s*\n"+
				"Тех. ключ: `%s`\n\n"+
				"Триггер добавлен в систему.\n"+
				"Теперь можно добавить паттерны и ответы.",
			safeMarkdown(triggerName),
			safeCode(techKey),
		)
	} else {
		resultText = fmt.Sprintf(
			"❌ *Ошибка создания триггера*\n\n"+
				"Название: *%s*\n"+
				"Тех. ключ: `%s`\n\n"+
				"Ошибка: %s",
			safeMarkdown(triggerName),
			safeCode(techKey),
			errorMsg,
		)
	}

	// Кнопки после создания
	detailCallback := fmt.Sprintf("admin:trigger:detail:%s", techKey)
	listCallback := "admin:triggers:list"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 К списку", listCallback),
		),
	)

	if success {
		// Добавляем кнопку открыть карточку если успешно
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔍 Открыть карточку", detailCallback),
				tgbotapi.NewInlineKeyboardButtonData("📋 К списку", listCallback),
			),
		)
	}

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

// cleanupNewTriggerStates - очистка устаревших состояний
func cleanupNewTriggerStates() {
	now := time.Now()
	for userID, state := range newTriggerAddStates {
		if now.Sub(state.CreatedAt) > 5*time.Minute {
			log.Printf("🧹 Очистка устаревшего состояния нового триггера для user_id: %d", userID)
			delete(newTriggerAddStates, userID)
		}
	}
}
