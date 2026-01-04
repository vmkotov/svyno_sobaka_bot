package mybot

import (
    "log"
    
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleMessage - основная функция обработки сообщения
// Координирует работу всех модулей
func HandleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, forwardChatID int64) {
    // Логируем что получили
    log.Printf("💬 Сообщение от @%s: %s", msg.From.UserName, msg.Text)
    
    // 1. Пересылаем сообщение (если нужно)
    forwardMessage(bot, msg, forwardChatID)
    
    // 2. Проверяем и обрабатываем команды
    if msg.IsCommand() {
        handleCommand(bot, msg)
    }
}
