package mybot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBDtechCallback - обработка callback'ов BDtech раздела
func HandleBDtechCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	if len(parts) < 3 {
		log.Printf("⚠️ Неполный bdtech callback: %v", parts)
		return
	}

	// parts[0] = "admin", parts[1] = "bdtech"
	switch parts[2] {
	case "menu":
		showBDtechMainMenu(bot, callbackQuery, db)
	case "tables":
		// Делегируем обработку UI_nav_menu_admin_BDtech_tables.go
		HandleBDtechTablesCallback(bot, callbackQuery, parts[3:])
	case "columns":
		// Делегируем обработку UI_nav_menu_admin_BDtech_columns.go
		HandleBDtechColumnsCallback(bot, callbackQuery, parts[3:])
	case "selects":
		// Делегируем обработку UI_nav_menu_admin_BDtech_selects.go
		HandleBDtechSelectsCallback(bot, callbackQuery, parts[3:])
	case "json":
		// Делегируем обработку UI_nav_menu_admin_BDtech_json.go
		HandleBDtechJSONCallback(bot, callbackQuery, parts[3:], db)
	case "procedures":
		// Делегируем обработку UI_nav_menu_admin_BDtech_procedures.go
		HandleBDtechProceduresCallback(bot, callbackQuery, parts[3:])
	case "functions":
		// Делегируем обработку UI_nav_menu_admin_BDtech_functions.go
		HandleBDtechFunctionsCallback(bot, callbackQuery, parts[3:])
	case "logs":
		// Делегируем обработку UI_nav_menu_admin_BDtech_logs.go
		HandleBDtechLogsCallback(bot, callbackQuery, parts[3:])
	default:
		log.Printf("⚠️ Неизвестный bdtech раздел: %s", parts[2])
	}
}

// showBDtechMainMenu показывает главное меню BDtech операций
func showBDtechMainMenu(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Получаем и анализируем структуру БД
	dbStats := getDatabaseStats(db)

	text := fmt.Sprintf("🛠️ *БД Тех операции*\n\n%s\n\nВыберите раздел:", dbStats)

	// Создаем inline-клавиатуру с кнопками в 3 колонки
	tablesBtn := tgbotapi.NewInlineKeyboardButtonData("📊 Таблицы", "admin:bdtech:tables:menu")
	columnsBtn := tgbotapi.NewInlineKeyboardButtonData("🗂️ Колонки", "admin:bdtech:columns:menu")
	selectsBtn := tgbotapi.NewInlineKeyboardButtonData("🔍 SELECTы", "admin:bdtech:selects:menu")
	
	jsonBtn := tgbotapi.NewInlineKeyboardButtonData("📄 JSON", "admin:bdtech:json:export")
	proceduresBtn := tgbotapi.NewInlineKeyboardButtonData("⚙️ Процедуры", "admin:bdtech:procedures:menu")
	functionsBtn := tgbotapi.NewInlineKeyboardButtonData("📞 Функции", "admin:bdtech:functions:menu")
	
	logsBtn := tgbotapi.NewInlineKeyboardButtonData("📝 Логи", "admin:bdtech:logs:menu")
	backBtn := tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "admin:menu")

	// Распределяем кнопки по рядам
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tablesBtn, columnsBtn, selectsBtn),
		tgbotapi.NewInlineKeyboardRow(jsonBtn, proceduresBtn, functionsBtn),
		tgbotapi.NewInlineKeyboardRow(logsBtn, backBtn),
	)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		text,
		inlineKeyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки меню BDtech: %v", err)
	}
}

// getDatabaseStats возвращает статистику по схеме svyno_sobaka_bot
func getDatabaseStats(db *sql.DB) string {
	if db == nil {
		return "❌ БД не подключена"
	}

	// Получаем полную структуру БД
	var jsonData string
	err := db.QueryRow("SELECT svyno_sobaka_bot.get_database_structure_complete()").Scan(&jsonData)
	if err != nil {
		log.Printf("❌ Ошибка получения структуры БД: %v", err)
		return "⚠️ Не удалось загрузить структуру БД"
	}

	// Парсим JSON
	var schemas []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &schemas); err != nil {
		log.Printf("❌ Ошибка парсинга JSON структуры БД: %v", err)
		return "⚠️ Ошибка формата данных БД"
	}

	// Ищем схему svyno_sobaka_bot
	var targetSchema map[string]interface{}
	var svynoSchemaFound bool
	
	for _, schema := range schemas {
		if schemaName, ok := schema["schema_name"].(string); ok && schemaName == "svyno_sobaka_bot" {
			targetSchema = schema
			svynoSchemaFound = true
			break
		}
	}

	if !svynoSchemaFound {
		return "⚠️ Схема `svyno_sobaka_bot` не найдена"
	}

	// Считаем статистику
	tables, _ := targetSchema["tables"].([]interface{})
	totalTables := len(tables)
	
	totalColumns := 0
	var tableNames []string
	
	for _, tableObj := range tables {
		if table, ok := tableObj.(map[string]interface{}); ok {
			if tableName, ok := table["table_name"].(string); ok {
				tableNames = append(tableNames, fmt.Sprintf("`%s`", tableName))
			}
			
			if columns, ok := table["columns"].([]interface{}); ok {
				totalColumns += len(columns)
			}
		}
	}

	// Форматируем вывод
	stats := fmt.Sprintf("📊 *Схема:* `svyno_sobaka_bot`\n"+
		"• Таблиц: %d\n"+
		"• Колонок: %d\n",
		totalTables, totalColumns)

	// Показываем первые 5 таблиц
	if len(tableNames) > 0 {
		displayTables := tableNames
		if len(tableNames) > 5 {
			displayTables = tableNames[:5]
			stats += fmt.Sprintf("\n📋 *Таблицы (первые 5):*\n%s\n... и ещё %d таблиц",
				strings.Join(displayTables, ", "), len(tableNames)-5)
		} else {
			stats += fmt.Sprintf("\n📋 *Таблицы:*\n%s", strings.Join(displayTables, ", "))
		}
	}

	return stats
}
