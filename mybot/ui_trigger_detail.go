package mybot

import (
    "fmt"
    "log"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GenerateTriggerDetailCard создает детальную карточку триггера
func GenerateTriggerDetailCard(trigger *Trigger, fromPage int) (string, tgbotapi.InlineKeyboardMarkup) {
    if trigger == nil {
        return createErrorMessage("unknown"), createBackButton(fromPage)
    }
    
    // Форматируем детали
    message := formatTriggerDetail(trigger)
    
    // Логируем сообщение для отладки
    log.Printf("🔍 Детальная карточка для %s, длина: %d байт", 
        trigger.TriggerName, len(message))
    
    // Проверим Markdown проблемы
    if strings.Count(message, "*")%2 != 0 {
        log.Printf("⚠️ Нечетное количество звёздочек в Markdown: %d", 
            strings.Count(message, "*"))
    }
    
    keyboard := createDetailKeyboard(trigger.TechKey, fromPage)
    
    return message, keyboard
}

// HandleTriggerDetailCallback обрабатывает callback детальной карточки
func HandleTriggerDetailCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string) {
    // Убираем "часики"
    callback := tgbotapi.NewCallback(callbackQuery.ID, "")
    bot.Request(callback)
    
    if len(parts) < 3 {
        log.Printf("⚠️ Неполный callback_data для деталей триггера: %v", parts)
        return
    }
    
    techKey := parts[2] // format: "trigger:detail:tech_key"
    
    // Извлекаем номер страницы из сообщения или используем 0
    fromPage := extractPageFromMessage(callbackQuery.Message.Text)
    
    // Получаем триггер
    trigger := GetTriggerByTechKey(techKey)
    
    // Генерируем детальную карточку
    message, keyboard := GenerateTriggerDetailCard(trigger, fromPage)
    
    // Отладочная информация
    log.Printf("📝 Отправляем сообщение длиной %d байт", len(message))
    
    // Редактируем сообщение
    msg := tgbotapi.NewEditMessageTextAndMarkup(
        callbackQuery.Message.Chat.ID,
        callbackQuery.Message.MessageID,
        message,
        keyboard,
    )
    msg.ParseMode = "Markdown"
    
    if _, err := bot.Send(msg); err != nil {
        log.Printf("❌ Ошибка отправки детальной карточки: %v", err)
        
        // Пробуем отправить без Markdown
        log.Printf("🔄 Пробуем отправить без Markdown...")
        msg.ParseMode = ""
        if _, err2 := bot.Send(msg); err2 != nil {
            log.Printf("❌ Ошибка даже без Markdown: %v", err2)
        } else {
            log.Printf("✅ Отправлено без Markdown")
        }
    }
}

// ================= ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ =================

func createErrorMessage(techKey string) string {
    return fmt.Sprintf("❌ Триггер с ключом `%s` не найден\n\n"+
        "Возможно, он был удален или изменен. "+
        "Используйте /refresh_me чтобы обновить список.", safeMarkdown(techKey))
}

func formatTriggerDetail(trigger *Trigger) string {
    // Форматируем паттерны с умным экранированием
    patternsText := formatPatterns(trigger.Patterns)
    
    // Форматируем ответы с умным экранированием
    responsesText := formatResponses(trigger.Responses)
    
    // Основное сообщение - используем safeMarkdown для текста
    return fmt.Sprintf(
        "🎯 *%s*\n\n"+
        "🔑 Тех. ключ: `%s`\n"+
        "🎯 Приоритет: %d\n"+
        "🎲 Вероятность: %d%%\n"+
        "📊 Паттернов: %d | Ответов: %d\n\n"+
        "🔍 *Паттерны:*\n%s\n\n"+
        "💬 *Ответы:*\n%s\n\n"+
        "#%s",
        safeMarkdown(trigger.TriggerName),           // Умное экранирование
        safeMarkdown(trigger.TechKey),               // Умное экранирование
        trigger.Priority,
        int(trigger.Probability*100),
        len(trigger.Patterns),
        len(trigger.Responses),
        patternsText,      // Уже экранировано в formatPatterns
        responsesText,     // Уже экранировано в formatResponses
        trigger.TechKey,   // Хештег без экранирования (Telegram сам разберется)
    )
}

func formatPatterns(patterns []Pattern) string {
    if len(patterns) == 0 {
        return "Нет паттернов"
    }
    
    var builder strings.Builder
    for i, p := range patterns {
        // Для паттернов внутри ` ` используем safeCode
        escapedPattern := safeMarkdown(p.PatternText)
        builder.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, escapedPattern))
    }
    return builder.String()
}

func formatResponses(responses []Response) string {
    if len(responses) == 0 {
        return "Нет ответов"
    }
    
    var builder strings.Builder
    for i, r := range responses {
        // Для ответов используем умное экранирование
        escapedResponse := safeMarkdown(r.ResponseText)
        builder.WriteString(fmt.Sprintf("%d. %s (вес: %d)\n", 
            i+1, escapedResponse, r.ResponseWeight))
    }
    return builder.String()
}

func createDetailKeyboard(techKey string, fromPage int) tgbotapi.InlineKeyboardMarkup {
    // Кнопка "Назад" возвращает на ту же страницу
    backCallback := fmt.Sprintf("triggers:page:%d", fromPage)
    
    // Кнопка "Главная"
    homeCallback := "menu:main"
    
    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", backCallback),
            tgbotapi.NewInlineKeyboardButtonData("🏠 Главная", homeCallback),
        ),
    )
    
    return keyboard
}

func createBackButton(fromPage int) tgbotapi.InlineKeyboardMarkup {
    backCallback := fmt.Sprintf("triggers:page:%d", fromPage)
    
    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", backCallback),
        ),
    )
    
    return keyboard
}

// extractPageFromMessage пытается извлечь номер страницы из текста сообщения
func extractPageFromMessage(text string) int {
    // Простая реализация - всегда возвращаем 0
    // TODO: можно добавить парсинг "Триггеры 1-10 из 50"
    return 0
}
