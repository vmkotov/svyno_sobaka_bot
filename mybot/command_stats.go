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

// UserStat - структура для парсинга JSON из функции get_svyno_sobaka_stats_by_chat_id
type UserStat struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ChatID    int64  `json:"chat_id"`
	ChatType  string `json:"chat_type"`
	Title     string `json:"title"`
	Cnt       int    `json:"cnt"`
}

// HandleStatsCommand - обработчик команды /stats
func HandleStatsCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, db *sql.DB) {
	log.Printf("📊 Команда /stats от @%s в чате %d", msg.From.UserName, msg.Chat.ID)

	// Проверяем подключение к БД
	if db == nil {
		log.Println("⚠️ БД не подключена, не могу получить статистику")
		SendMessage(bot, msg.Chat.ID, "❌ БД не подключена", "ошибка stats")
		return
	}

	// Получаем ID чата из сообщения
	chatID := msg.Chat.ID
	log.Printf("🔍 Запрашиваем статистику для чата %d", chatID)

	// Вызываем SQL-функцию
	var jsonData string
	err := db.QueryRow("SELECT svyno_sobaka_bot.get_svyno_sobaka_stats_by_chat_id($1)", chatID).Scan(&jsonData)
	if err != nil {
		log.Printf("❌ Ошибка вызова SQL-функции: %v", err)
		SendMessage(bot, msg.Chat.ID, "❌ Ошибка получения статистики из БД", "ошибка stats")
		return
	}

	log.Printf("✅ Получены данные из БД: %d байт", len(jsonData))

	// Парсим JSON в массив структур
	var stats []UserStat
	if err := json.Unmarshal([]byte(jsonData), &stats); err != nil {
		log.Printf("❌ Ошибка парсинга JSON: %v", err)
		SendMessage(bot, msg.Chat.ID, "❌ Ошибка обработки данных статистики", "ошибка stats")
		return
	}

	// Проверяем, есть ли данные
	if len(stats) == 0 {
		log.Printf("📭 Нет статистики для чата %d", chatID)
		SendMessage(bot, msg.Chat.ID, "📊 В этом чате пока никто не был свинособакой.", "stats пусто")
		return
	}

	log.Printf("📊 Найдено %d записей статистики", len(stats))

	// Форматируем и отправляем сообщение
	messageText := formatStatsMessage(stats, chatID)
	SendMessage(bot, msg.Chat.ID, messageText, "stats")
}

// formatStatsMessage - форматирует статистику в читаемое сообщение
func formatStatsMessage(stats []UserStat, chatID int64) string {
	// Сортируем по убыванию количества
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Cnt > stats[j].Cnt
	})

	// Определяем название чата
	chatTitle := "этом чате"
	if len(stats) > 0 {
		chatTitle = stats[0].Title
	}

	// Формируем заголовок
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📊 *Статистика свинособак в \"%s\":*\n\n", escapeMarkdown(chatTitle)))

	// Определяем лимит показа (не больше 15, чтобы не превысить лимит сообщения)
	limit := 15
	if len(stats) < limit {
		limit = len(stats)
	}

	// Формируем строки статистики
	for i := 0; i < limit; i++ {
		stat := stats[i]
		
		// Формируем имя пользователя по правилам: @username → Имя Фамилия → Имя → ID
		userName := formatUserName(stat)
		
		// Добавляем эмодзи для топ-3
		emoji := ""
		switch i {
		case 0:
			emoji = "🥇 "
		case 1:
			emoji = "🥈 "
		case 2:
			emoji = "🥉 "
		}
		
		// Добавляем строку
		builder.WriteString(fmt.Sprintf("%s%s — *%d* раз\n", 
			emoji, 
			escapeMarkdown(userName), 
			stat.Cnt))
	}

	// Если записей больше лимита, добавляем информацию
	if len(stats) > limit {
		builder.WriteString(fmt.Sprintf("\n*... и ещё %d пользователей*", len(stats)-limit))
	}

	return builder.String()
}

// formatUserName - форматирует имя пользователя по правилам:
// @username (если есть) → Имя Фамилия (если есть) → Имя → ID
func formatUserName(stat UserStat) string {
	// Если есть username (и он не пустой)
	if stat.Username != "" {
		return "@" + stat.Username
	}
	
	// Если есть имя и фамилия
	if stat.FirstName != "" && stat.LastName != "" {
		return stat.FirstName + " " + stat.LastName
	}
	
	// Если есть только имя
	if stat.FirstName != "" {
		return stat.FirstName
	}
	
	// Если ничего нет, возвращаем ID (хотя в JSON всегда есть first_name)
	return fmt.Sprintf("ID: %d", stat.ChatID) // fallback
}
