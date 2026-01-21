// ============================================================================
// ФАЙЛ: admin_ui.go
// UI компоненты для админ-панели
// ============================================================================
package mybot

import (
    "fmt"
    "log"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendAdminMainMenu отправляет главное меню админки "СвиноАдминка"
// Показывается только администраторам при команде /start
func SendAdminMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
    text := fmt.Sprintf(
        "🐷 *СвиноАдминка*\n\n" +
        "Выберите действие:",
    )

    // Создаем inline-клавиатуру с двумя кнопками
    refreshButton := tgbotapi.NewInlineKeyboardButtonData(
        "🔄 Обновить триггеры", 
        "admin:refresh",
    )
    triggersButton := tgbotapi.NewInlineKeyboardButtonData(
        "📋 Триггеры", 
        "admin:triggers:list",
    )

    // Две кнопки в один ряд
    inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(refreshButton, triggersButton),
    )
    
    // Отправляем сообщение
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ReplyMarkup = inlineKeyboard
    msg.ParseMode = "Markdown"
    
    if _, err := bot.Send(msg); err != nil {
        log.Printf("❌ Ошибка отправки меню админки: %v", err)
    } else {
        log.Printf("✅ Меню админки отправлено в чат %d", chatID)
    }
}

// SendUserMainMenu отправляет меню для обычных пользователей
// Показывается при команде /start для не-администраторов
