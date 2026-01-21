package mybot

import (
    "log"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleBDtechCallback - обработка callback'ов BDtech раздела
func HandleBDtechCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string) {
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
        showBDtechMainMenu(bot, callbackQuery)
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
        HandleBDtechJSONCallback(bot, callbackQuery, parts[3:])
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
func showBDtechMainMenu(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
    text := "🛠️ *БД Тех операции*\n\nВыберите раздел:"

    // Создаем inline-клавиатуру с кнопками в 3 колонки
    tablesBtn := tgbotapi.NewInlineKeyboardButtonData("📊 Таблицы", "admin:bdtech:tables:menu")
    columnsBtn := tgbotapi.NewInlineKeyboardButtonData("🗂️ Колонки", "admin:bdtech:columns:menu")
    selectsBtn := tgbotapi.NewInlineKeyboardButtonData("🔍 SELECTы", "admin:bdtech:selects:menu")
    
    jsonBtn := tgbotapi.NewInlineKeyboardButtonData("📄 JSON", "admin:bdtech:json:menu")
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
