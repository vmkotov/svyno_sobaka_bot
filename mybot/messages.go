package mybot

import (
    "fmt"
    "log"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStartCommand - обработка команды /start
func handleStartCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    name := msg.From.FirstName
    if msg.From.UserName != "" {
        name = "@" + msg.From.UserName
    }
    
    text := fmt.Sprintf(
        "привет, я Свинособака. ты, %s, кстати тоже!\n" +
        "ждём от Грека БТ, ФТ, ТЗ и прочую хуйню.\n" +
        "а пока иди нахуй", 
        name)
    
    sendMessage(bot, msg.Chat.ID, text, "старт")
}

// handleHelpCommand - обработка команды /help
func handleHelpCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    text := "📋 Команды:\n/start - Начать\n/help - Помощь"
    sendMessage(bot, msg.Chat.ID, text, "справка")
}

// sendMessage - общая функция отправки сообщений
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text, context string) {
    reply := tgbotapi.NewMessage(chatID, text)
    
    if _, err := bot.Send(reply); err != nil {
        log.Printf("❌ Ошибка отправки %s: %v", context, err)
    } else {
        log.Printf("✅ Отправлен %s", context)
    }
}
