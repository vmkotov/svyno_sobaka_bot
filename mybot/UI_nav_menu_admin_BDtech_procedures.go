package mybot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Константы для меню процедур
const (
	proceduresPerPage      = 10
	procedureMaxNameLength = 30
)

// HandleBDtechProceduresCallback - обработка раздела процедур
func HandleBDtechProceduresCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	if len(parts) == 0 || parts[0] == "menu" {
		showProceduresList(bot, callbackQuery, db, 0)
		return
	}

	switch parts[0] {
	case "page":
		// Обработка пагинации
		if len(parts) >= 2 {
			page := 0
			if n, err := fmt.Sscanf(parts[1], "%d", &page); err == nil && n == 1 {
				showProceduresList(bot, callbackQuery, db, page)
				return
			}
		}
		showProceduresList(bot, callbackQuery, db, 0)

	case "view":
		// Просмотр конкретной процедуры
		// Новый формат: admin:proc:view:schema:procedureName
		if len(parts) >= 5 {
			schema := parts[3]
			procedureName := parts[4]
			viewProcedureCode(bot, callbackQuery, db, schema, procedureName)
			return
		}
		showProceduresList(bot, callbackQuery, db, 0)

	default:
		showProceduresList(bot, callbackQuery, db, 0)
	}
}

// showProceduresList показывает список процедур схемы svyno_sobaka_bot
func showProceduresList(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, page int) {
	if db == nil {
		text := "❌ БД не подключена"
		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
	}

	// Получаем список процедур
	var jsonData string
	err := db.QueryRow("SELECT svyno_sobaka_bot.get_database_functions_procedures()").Scan(&jsonData)
	if err != nil {
		log.Printf("❌ Ошибка получения процедур: %v", err)
		text := "❌ Не удалось загрузить список процедур"
		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
	}

	// Парсим JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		log.Printf("❌ Ошибка парсинга JSON процедур: %v", err)
		text := "❌ Ошибка формата данных процедур"
		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
	}

	// Извлекаем массив процедур
	functionsArray, ok := result["functions_and_procedures"].([]interface{})
	if !ok {
		text := "❌ Нет данных о процедурах"
		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
	}

	// Фильтруем только схему svyno_sobaka_bot
	var procedures []map[string]interface{}
	for _, item := range functionsArray {
		if proc, ok := item.(map[string]interface{}); ok {
			if schema, ok := proc["schema"].(string); ok && schema == "svyno_sobaka_bot" {
				procedures = append(procedures, proc)
			}
		}
	}

	if len(procedures) == 0 {
		text := "📋 *БД Тех - Процедуры схемы svyno_sobaka_bot*\n\n" +
			"В схеме нет процедур или функций"

		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
	}

	// Сортируем по алфавиту
	sort.Slice(procedures, func(i, j int) bool {
		nameI, _ := procedures[i]["procedure_name"].(string)
		nameJ, _ := procedures[j]["procedure_name"].(string)
		return strings.ToLower(nameI) < strings.ToLower(nameJ)
	})

	totalProcedures := len(procedures)
	startIdx := page * proceduresPerPage
	endIdx := startIdx + proceduresPerPage

	// Проверяем границы
	if startIdx >= totalProcedures {
		startIdx = 0
		page = 0
		endIdx = proceduresPerPage
	}
	if endIdx > totalProcedures {
		endIdx = totalProcedures
	}

	// Формируем текст
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("⚙️ *БД Тех - Процедуры схемы svyno_sobaka_bot*\n"))
	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━\n")
	builder.WriteString(fmt.Sprintf("Всего: %d\n\n", totalProcedures))

	// Создаем кнопки для процедур текущей страницы
	var buttonRows [][]tgbotapi.InlineKeyboardButton

	for i := startIdx; i < endIdx; i++ {
		proc := procedures[i]
		procName, _ := proc["procedure_name"].(string)
		procType, _ := proc["type"].(string)
		procNum := i + 1

		// Форматируем текст кнопки
		buttonText := formatProcedureButton(procName, procType, procNum)

		// Создаем callback_data с проверкой длины (макс 64 байта в Telegram)
		const shortPrefix = "admin:proc:"
		schema := "svyno_sobaka_bot"
		callbackData := fmt.Sprintf("%sview:%s:%s", shortPrefix, schema, procName)

		// Логируем для отладки
		log.Printf("📏 Callback для %s.%s: %d символов (макс: 64)", schema, procName, len(callbackData))
		if len(callbackData) > 64 {
			// Укорачиваем имя процедуры
			prefixLength := len(shortPrefix) + len("view:") + len(schema) + 1
			maxProcNameLength := 64 - prefixLength
			if maxProcNameLength > 0 && len(procName) > maxProcNameLength {
				shortName := procName[:maxProcNameLength]
				callbackData = fmt.Sprintf("%sview:%s:%s", shortPrefix, schema, shortName)
			}
		}

		// Создаем кнопку
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Добавляем навигацию
	navRow := createProceduresNavigationButtons(page, totalProcedures)
	if len(navRow) > 0 {
		buttonRows = append(buttonRows, navRow)
	}

	// Добавляем кнопку возврата
	backBtn := tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в BDtech", "admin:bdtech:menu")
	buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(backBtn))

	// Формируем финальное сообщение
	pageInfo := ""
	if totalProcedures > proceduresPerPage {
		pageInfo = fmt.Sprintf("\nСтраница %d/%d", page+1, (totalProcedures+proceduresPerPage-1)/proceduresPerPage)
	}

	finalText := fmt.Sprintf("⚙️ *БД Тех - Процедуры схемы svyno_sobaka_bot*\n"+
		"━━━━━━━━━━━━━━━━━━━━━━━━\n"+
		"Всего процедур/функций: %d%s\n\n"+
		"Нажмите на процедуру чтобы получить её SQL код:",
		totalProcedures, pageInfo)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		finalText,
		tgbotapi.NewInlineKeyboardMarkup(buttonRows...),
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки списка процедур: %v", err)
	}
}

// formatProcedureButton форматирует текст для кнопки процедуры
func formatProcedureButton(name, procType string, number int) string {
	// Обрезаем название если нужно
	displayName := name
	if len(displayName) > procedureMaxNameLength {
		displayName = displayName[:procedureMaxNameLength-3] + "..."
	}

	// Формируем текст кнопки
	buttonText := fmt.Sprintf("%d. %s", number, displayName)

	// Добавляем тип если помещается
	if len(buttonText) < 25 { // Примерная проверка на длину
		typeSymbol := ""
		switch procType {
		case "PROCEDURE":
			typeSymbol = " 🅿️"
		case "FUNCTION":
			typeSymbol = " 🅵"
		case "AGGREGATE":
			typeSymbol = " 🅰️"
		}
		if typeSymbol != "" {
			buttonText += typeSymbol
		}
	}

	return buttonText
}

// createProceduresNavigationButtons создает кнопки навигации для процедур
func createProceduresNavigationButtons(currentPage, totalProcedures int) []tgbotapi.InlineKeyboardButton {
	var buttons []tgbotapi.InlineKeyboardButton

	// Рассчитываем общее количество страниц
	totalPages := (totalProcedures + proceduresPerPage - 1) / proceduresPerPage

	// Определяем, какие кнопки показывать
	hasPrevPage := currentPage > 0
	hasNextPage := (currentPage + 1) < totalPages

	// Кнопка "Назад" (если не первая страница)
	if hasPrevPage {
		callbackData := fmt.Sprintf("admin:bdtech:procedures:page:%d", currentPage-1)
		button := tgbotapi.NewInlineKeyboardButtonData("⏪ Назад", callbackData)
		buttons = append(buttons, button)
	}

	// Кнопка "Главная страница" (если есть пагинация)
	if totalPages > 1 {
		callbackData := "admin:bdtech:procedures:menu"
		button := tgbotapi.NewInlineKeyboardButtonData("📄 Страница 1", callbackData)
		buttons = append(buttons, button)
	}

	// Кнопка "Далее" (если не последняя страница)
	if hasNextPage {
		callbackData := fmt.Sprintf("admin:bdtech:procedures:page:%d", currentPage+1)
		button := tgbotapi.NewInlineKeyboardButtonData("⏩ Далее", callbackData)
		buttons = append(buttons, button)
	}

	return buttons
}

// viewProcedureCode отправляет SQL код процедуры как файл
func viewProcedureCode(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, schema, procedureName string) {
	if db == nil {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ БД не подключена")
		bot.Request(callback)
		return
	}

	log.Printf("📄 Запрос кода процедуры: %s.%s от @%s",
		schema, procedureName, callbackQuery.From.UserName)

	// Получаем данные о процедуре
	var jsonData string
	err := db.QueryRow("SELECT svyno_sobaka_bot.get_database_functions_procedures()").Scan(&jsonData)
	if err != nil {
		log.Printf("❌ Ошибка получения процедур: %v", err)
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Ошибка получения данных")
		bot.Request(callback)
		return
	}

	// Парсим JSON и ищем нужную процедуру
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		log.Printf("❌ Ошибка парсинга JSON: %v", err)
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Ошибка данных")
		bot.Request(callback)
		return
	}

	// Ищем процедуру
	functionsArray, ok := result["functions_and_procedures"].([]interface{})
	if !ok {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Процедура не найдена")
		bot.Request(callback)
		return
	}

	var targetProc map[string]interface{}
	for _, item := range functionsArray {
		if proc, ok := item.(map[string]interface{}); ok {
			procSchema, _ := proc["schema"].(string)
			procName, _ := proc["procedure_name"].(string)
			if procSchema == schema && procName == procedureName {
				targetProc = proc
				break
			}
		}
	}

	if targetProc == nil {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Процедура не найдена")
		bot.Request(callback)
		return
	}

	// Извлекаем код процедуры
	procedureCode, ok := targetProc["procedure_code"].(string)
	if !ok || procedureCode == "" {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Нет кода процедуры")
		bot.Request(callback)
		return
	}

	procType, _ := targetProc["type"].(string)

	// Формируем файл с заголовком
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	header := fmt.Sprintf(`-- Schema: %s
-- Procedure: %s
-- Type: %s
-- Generated: %s
-- 
-- Original SQL code:
--

`, schema, procedureName, procType, timestamp)

	fullCode := header + procedureCode

	// Создаем имя файла
	fileName := fmt.Sprintf("%s.%s.txt", schema, procedureName)

	// Отправляем как файл
	file := tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: []byte(fullCode),
	}

	msg := tgbotapi.NewDocument(callbackQuery.Message.Chat.ID, file)
	msg.Caption = fmt.Sprintf("📄 %s.%s\nТип: %s\nРазмер: %.2f KB",
		schema, procedureName, procType, float64(len(fullCode))/1024)

	// Убираем "часики" у callback
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	bot.Request(callback)

	// Отправляем файл
	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки файла процедуры: %v", err)
		callback := tgbotapi.NewCallback(callbackQuery.ID, "❌ Ошибка отправки файла")
		bot.Request(callback)
	} else {
		log.Printf("✅ Файл процедуры отправлен: %s", fileName)
	}
}
