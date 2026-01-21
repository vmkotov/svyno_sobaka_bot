package mybot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBDtechJSONCallback - обработка раздела JSON операций
func HandleBDtechJSONCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	if len(parts) == 0 {
		showJSONMenu(bot, callbackQuery, db)
		return
	}

	switch parts[0] {
	case "export":
		exportDatabaseJSON(bot, callbackQuery, db)
	case "menu":
		showJSONMenu(bot, callbackQuery, db)
	default:
		showJSONMenu(bot, callbackQuery, db)
	}
}

// showJSONMenu показывает меню JSON операций
func showJSONMenu(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	text := "📄 *БД Тех - JSON операции*\n\nВыберите действие:"

	exportBtn := tgbotapi.NewInlineKeyboardButtonData("📥 Экспорт структуры", "admin:bdtech:json:export")
	backBtn := tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "admin:bdtech:menu")

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(exportBtn),
		tgbotapi.NewInlineKeyboardRow(backBtn),
	)

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		text,
		inlineKeyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки меню JSON: %v", err)
	}
}

// exportDatabaseJSON экспортирует структуру схемы svyno_sobaka_bot в JSON файл
func exportDatabaseJSON(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	if db == nil {
		sendErrorMessage(bot, callbackQuery, "❌ БД не подключена")
		return
	}

	// Уведомляем о начале процесса
	processingMsg := tgbotapi.NewEditMessageText(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		"⏳ *Экспорт структуры БД...*\n\nПодготавливаю JSON файл...",
	)
	processingMsg.ParseMode = "Markdown"
	bot.Send(processingMsg)

	// Получаем полную структуру БД
	var fullJSON string
	err := db.QueryRow("SELECT svyno_sobaka_bot.get_database_structure_complete()").Scan(&fullJSON)
	if err != nil {
		sendErrorMessage(bot, callbackQuery, fmt.Sprintf("❌ Ошибка получения структуры БД: %v", err))
		return
	}

	// Парсим JSON для фильтрации
	var schemas []map[string]interface{}
	if err := json.Unmarshal([]byte(fullJSON), &schemas); err != nil {
		sendErrorMessage(bot, callbackQuery, fmt.Sprintf("❌ Ошибка парсинга JSON: %v", err))
		return
	}

	// Фильтруем: оставляем только схему svyno_sobaka_bot
	var filteredSchemas []map[string]interface{}
	for _, schema := range schemas {
		if schemaName, ok := schema["schema_name"].(string); ok && schemaName == "svyno_sobaka_bot" {
			filteredSchemas = append(filteredSchemas, schema)
			break
		}
	}

	if len(filteredSchemas) == 0 {
		sendErrorMessage(bot, callbackQuery, "❌ Схема `svyno_sobaka_bot` не найдена")
		return
	}

	// Преобразуем обратно в JSON
	filteredJSON, err := json.MarshalIndent(filteredSchemas, "", "  ")
	if err != nil {
		sendErrorMessage(bot, callbackQuery, fmt.Sprintf("❌ Ошибка форматирования JSON: %v", err))
		return
	}

	// Создаем временный файл
	tmpfile, err := ioutil.TempFile("", "svyno_sobaka_bot_*.json")
	if err != nil {
		sendErrorMessage(bot, callbackQuery, fmt.Sprintf("❌ Ошибка создания файла: %v", err))
		return
	}
	defer os.Remove(tmpfile.Name())

	// Записываем JSON в файл
	if _, err := tmpfile.Write(filteredJSON); err != nil {
		sendErrorMessage(bot, callbackQuery, fmt.Sprintf("❌ Ошибка записи в файл: %v", err))
		return
	}
	tmpfile.Close()

	// Формируем имя файла с датой
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("svyno_sobaka_bot_structure_%s.json", timestamp)

	// Отправляем файл (БЕЗ Markdown в подписи - используем обычный текст)
	fileMsg := tgbotapi.NewDocument(callbackQuery.Message.Chat.ID, tgbotapi.FilePath(tmpfile.Name()))
	
	// Формируем подпись БЕЗ Markdown форматирования
	caption := fmt.Sprintf("📄 Экспорт структуры БД\n\n"+
		"Схема: svyno_sobaka_bot\n"+
		"Дата: %s\n"+
		"Размер: %.2f KB\n\n"+
		"Файл: %s",
		time.Now().Format("02.01.2006 15:04:05"),
		float64(len(filteredJSON))/1024,
		filename)
	
	// Экранируем специальные символы для безопасности
	caption = escapeMarkdownV2(caption)
	fileMsg.Caption = caption
	fileMsg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(fileMsg); err != nil {
		log.Printf("❌ Ошибка отправки файла: %v", err)
		// Пробуем без Markdown
		fileMsg.ParseMode = ""
		fileMsg.Caption = strings.ReplaceAll(caption, "\\", "")
		if _, err2 := bot.Send(fileMsg); err2 != nil {
			sendErrorMessage(bot, callbackQuery, fmt.Sprintf("❌ Ошибка отправки файла: %v", err2))
			return
		}
	}

	// Обновляем оригинальное сообщение
	successMsg := tgbotapi.NewEditMessageText(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		fmt.Sprintf("✅ *Экспорт завершен!*\n\nФайл `%s` отправлен\\.\nРазмер: %.2f KB", 
			filename, float64(len(filteredJSON))/1024),
	)
	successMsg.ParseMode = "MarkdownV2"
	bot.Send(successMsg)

	log.Printf("✅ Экспорт JSON отправлен пользователю @%s", callbackQuery.From.UserName)
}

// sendErrorMessage отправляет сообщение об ошибке
func sendErrorMessage(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, message string) {
	// Используем MarkdownV2 с экранированием
	msg := tgbotapi.NewEditMessageText(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		escapeMarkdownV2(message),
	)
	msg.ParseMode = "MarkdownV2"
	
	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки сообщения об ошибке: %v", err)
	}
}

// escapeMarkdownV2 экранирует специальные символы для MarkdownV2
func escapeMarkdownV2(text string) string {
	// Список символов, которые нужно экранировать в MarkdownV2
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	
	return result
}
