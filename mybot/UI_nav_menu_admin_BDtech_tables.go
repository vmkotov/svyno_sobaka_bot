package mybot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBDtechTablesCallback - обработка раздела таблиц
func HandleBDtechTablesCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	if len(parts) == 0 || parts[0] == "menu" {
		showTablesList(bot, callbackQuery, db)
		return
	}
	
	// Для будущего расширения - детальная информация о таблице
	showTablesList(bot, callbackQuery, db)
}

// showTablesList показывает список таблиц схемы svyno_sobaka_bot
func showTablesList(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
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

	// Получаем структуру БД
	var jsonData string
	err := db.QueryRow("SELECT svyno_sobaka_bot.get_database_structure_complete()").Scan(&jsonData)
	if err != nil {
		log.Printf("❌ Ошибка получения структуры БД: %v", err)
		text := "❌ Не удалось загрузить структуру БД"
		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
	}

	// Парсим JSON
	var schemas []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &schemas); err != nil {
		log.Printf("❌ Ошибка парсинга JSON структуры БД: %v", err)
		text := "❌ Ошибка формата данных БД"
		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
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
		text := "❌ Схема svyno_sobaka_bot не найдена"
		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
	}

	// Получаем таблицы
	tables, ok := targetSchema["tables"].([]interface{})
	if !ok {
		text := "❌ Нет данных о таблицах"
		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
	}

	// Собираем информацию о таблицах
	type TableInfo struct {
		Name    string
		Columns int
		Comment string
	}

	var tablesInfo []TableInfo
	for _, tableObj := range tables {
		if table, ok := tableObj.(map[string]interface{}); ok {
			tableName, hasName := table["table_name"].(string)
			columns, hasColumns := table["columns"].([]interface{})
			tableComment, _ := table["table_comment"].(string)
			
			if hasName && hasColumns {
				// Если комментарий есть, берем первое предложение
				shortComment := ""
				if tableComment != "" {
					// Берем первое предложение до точки
					if idx := strings.Index(tableComment, "."); idx != -1 {
						shortComment = strings.TrimSpace(tableComment[:idx+1])
					} else {
						shortComment = tableComment
					}
				}
				
				tablesInfo = append(tablesInfo, TableInfo{
					Name:    tableName,
					Columns: len(columns),
					Comment: shortComment,
				})
			}
		}
	}

	if len(tablesInfo) == 0 {
		text := "📊 БД Тех - Таблицы схемы svyno_sobaka_bot\n\n" +
			"В схеме нет таблиц"
		
		msg := tgbotapi.NewEditMessageText(
			callbackQuery.Message.Chat.ID,
			callbackQuery.Message.MessageID,
			text,
		)
		bot.Send(msg)
		return
	}

	// Сортируем по алфавиту
	sort.Slice(tablesInfo, func(i, j int) bool {
		return strings.ToLower(tablesInfo[i].Name) < strings.ToLower(tablesInfo[j].Name)
	})

	// Формируем текст
	var builder strings.Builder
	builder.WriteString("📊 БД Тех - Таблицы схемы svyno_sobaka_bot\n")
	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━\n")
	builder.WriteString(fmt.Sprintf("Всего таблиц: %d\n\n", len(tablesInfo)))

	for i, table := range tablesInfo {
		// 1. **messages_log** [14 полей]. Логи сообщений
		builder.WriteString(fmt.Sprintf("%d. **%s** [%d полей]", 
			i+1, table.Name, table.Columns))
		
		if table.Comment != "" {
			builder.WriteString(fmt.Sprintf(". %s", table.Comment))
		}
		
		builder.WriteString("\n")
	}

	builder.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Кнопка возврата
	backBtn := tgbotapi.NewInlineKeyboardButtonData("🔙 Назад в BDtech", "admin:bdtech:menu")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(backBtn),
	)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		builder.String(),
		inlineKeyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки списка таблиц: %v", err)
		// Пробуем без Markdown
		msg.ParseMode = ""
		msg.Text = strings.ReplaceAll(builder.String(), "**", "")
		bot.Send(msg)
	}
}
