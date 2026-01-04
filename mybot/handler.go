package mybot

import (
    "database/sql"
    "log"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleMessage - обрабатывает сообщение
func HandleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, 
                   forwardChatID int64, db *sql.DB, botUsername string) {
    
    log.Printf("💬 Сообщение от @%s: %s", msg.From.UserName, msg.Text)
    
    // 1. Сохраняем в БД (если подключена)
    if db != nil {
        SaveMessageToDB(db, botUsername, msg)
    }
    
    // 2. Пересылаем
    forwardMessage(bot, msg, forwardChatID)
    
    // 3. Команды
    if msg.IsCommand() {
        handleCommand(bot, msg)
    }
}
