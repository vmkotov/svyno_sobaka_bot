package mybot

import (
    "fmt"
    "log"
    "strings"
    "time"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendMessageLog отправляет форматированный лог сообщения в указанный чат
func SendMessageLog(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, botUsername string, botID int64) {
    // ID чата для логов (фиксированный)
    logChatID := int64(-1003516004835)
    
    // Формируем лог
    logText := formatMessageLog(msg, botUsername, botID)
    
    // Создаем сообщение для отправки
    logMsg := tgbotapi.NewMessage(logChatID, logText)
    logMsg.ParseMode = "Markdown"
    logMsg.DisableWebPagePreview = true
    
    // Отправляем лог
    if _, err := bot.Send(logMsg); err != nil {
        log.Printf("❌ Ошибка отправки лога: %v", err)
    } else {
        log.Printf("✅ Лог отправлен в чат %d", logChatID)
    }
}

// formatMessageLog формирует форматированный текст лога
func formatMessageLog(msg *tgbotapi.Message, botUsername string, botID int64) string {
    var builder strings.Builder
    
    // Заголовок с временем
    messageTime := time.Unix(int64(msg.Date), 0)
    builder.WriteString(fmt.Sprintf("🤖 *Лог сообщения* `%s`\n\n", 
        messageTime.Format("15:04:05")))
    
    // Информация о чате
    chatTitle := getValueOrDefault(msg.Chat.Title, "не указано")
    builder.WriteString(fmt.Sprintf("💬 *Чат:* %s\n", escapeMarkdown(chatTitle)))
    
    // Тип чата (перевод на русский)
    chatType := translateChatType(msg.Chat.Type)
    builder.WriteString(fmt.Sprintf("📌 *Тип:* %s\n", chatType))
    builder.WriteString(fmt.Sprintf("🆔 *ID:* `%d`\n\n", msg.Chat.ID))
    
    // Информация о пользователе
    if msg.From != nil {
        // Полное имя
        fullName := strings.TrimSpace(fmt.Sprintf("%s %s", 
            getValueOrDefault(msg.From.FirstName, ""),
            getValueOrDefault(msg.From.LastName, "")))
        if fullName == "" {
            fullName = "не указано"
        }
        builder.WriteString(fmt.Sprintf("👤 *Пользователь:* %s\n", escapeMarkdown(fullName)))
        
        // Только имя
        firstName := getValueOrDefault(msg.From.FirstName, "не указано")
        builder.WriteString(fmt.Sprintf("📛 *Имя:* %s\n", escapeMarkdown(firstName)))
        
        // Username
        username := getValueOrDefault(msg.From.UserName, "не указано")
        if username != "не указано" {
            builder.WriteString(fmt.Sprintf("👤 *@%s*\n", escapeMarkdown(username)))
        }
        
        builder.WriteString(fmt.Sprintf("🆔 *ID:* `%d`\n\n", msg.From.ID))
    }
    
    // Текст сообщения или подпись
    messageText := getMessageText(msg)
    builder.WriteString(fmt.Sprintf("📝 *Сообщение:*\n```\n%s\n```\n\n", messageText))
    
    // Информация о боте
    builder.WriteString(fmt.Sprintf("🤖 *Информация о боте:*\n"))
    builder.WriteString(fmt.Sprintf("Бот: @%s\n", botUsername))
    builder.WriteString(fmt.Sprintf("Bot ID: `%d`", botID))
    
    return builder.String()
}

// getMessageText возвращает текст сообщения или подпись к медиа
func getMessageText(msg *tgbotapi.Message) string {
    if msg.Text != "" {
        return msg.Text
    }
    if msg.Caption != "" {
        return msg.Caption
    }
    if msg.Photo != nil {
        return "[Фото]"
    }
    if msg.Video != nil {
        return "[Видео]"
    }
    if msg.Document != nil {
        return "[Документ]"
    }
    if msg.Audio != nil {
        return "[Аудио]"
    }
    if msg.Voice != nil {
        return "[Голосовое сообщение]"
    }
    if msg.Sticker != nil {
        return "[Стикер]"
    }
    if msg.Location != nil {
        return "[Местоположение]"
    }
    if msg.Contact != nil {
        return "[Контакт]"
    }
    
    return "[Сообщение без текста]"
}

// translateChatType переводит тип чата на русский
func translateChatType(chatType string) string {
    switch chatType {
    case "supergroup":
        return "супергруппа"
    case "group":
        return "группа"
    case "private":
        return "личный чат"
    case "channel":
        return "канал"
    default:
        return chatType
    }
}

// getValueOrDefault возвращает значение или значение по умолчанию
func getValueOrDefault(value, defaultValue string) string {
    if value == "" {
        return defaultValue
    }
    return value
}

// escapeMarkdown экранирует символы Markdown
func escapeMarkdown(text string) string {
    // Экранируем специальные символы Markdown
    specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
    result := text
    for _, char := range specialChars {
        result = strings.ReplaceAll(result, char, "\\"+char)
    }
    return result
}
