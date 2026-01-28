// ============================================================================
// ФАЙЛ: UI_nav_menu_admin.go
// Обработка UI callback админки (admin:*)
// ============================================================================
package mybot

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleAdminUICallback - обработка UI callback админки
func HandleAdminUICallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	// Убираем "часики"
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("⚠️ Ошибка AnswerCallbackQuery: %v", err)
	}

	if len(parts) < 2 {
		log.Printf("⚠️ Неполный admin callback_data: %v", parts)
		return
	}

	switch parts[1] {
	case "menu":
		log.Printf("👑 Админское меню от @%s", callbackQuery.From.UserName)
		showAdminMenu(bot, callbackQuery)
	case "refresh":
		log.Printf("👑 Админское обновление триггеров от @%s", callbackQuery.From.UserName)
		handleAdminRefreshTriggers(bot, callbackQuery, db)
	case "triggers":
		handleAdminTriggersUICallback(bot, callbackQuery, parts, db)
	case "trigger":
		// Обработка нового триггера (admin:trigger:new)
		if len(parts) >= 3 && parts[2] == "new" {
			if len(parts) >= 4 && parts[3] == "cancel" {
				handleAddNewTriggerCancel(bot, callbackQuery)
			} else {
				handleAddNewTrigger(bot, callbackQuery)
			}
			return
		}

		// Обработка остальных кнопок триггера (нужно >=5 частей)
		if len(parts) >= 5 {
			switch parts[2] {
			case "pattern":
				if parts[3] == "add" {
					handleAddPattern(bot, callbackQuery, parts[4]) // techKey
					return
				}
				if parts[3] == "cancel" {
					handleAddPatternCancel(bot, callbackQuery, parts[4])
					return
				}
			case "response":
				if parts[3] == "add" {
					handleAddResponse(bot, callbackQuery, parts[4])
					return
				}
			case "prob":
				if parts[3] == "edit" {
					handleEditProbability(bot, callbackQuery, parts[4])
					return
				}
			}
		}

		// Если не новые кнопки, то это детальная карточка
		// admin:trigger:detail:TECH_KEY (должно быть 4 части)
		if len(parts) >= 4 && parts[2] == "detail" {
			HandleAdminTriggerDetailCallback(bot, callbackQuery, parts, db)
			return
		}

		log.Printf("⚠️ Неизвестный trigger callback: %v", parts)

	case "bdtech":
	case "proc":
		log.Printf("⚙️ Обработка процедур от @%s", callbackQuery.From.UserName)
		HandleBDtechCallback(bot, callbackQuery, parts, db)
	case "home":
		log.Printf("🏠 Главная из админки от @%s", callbackQuery.From.UserName)
		EditUserMenu(bot, callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
	default:
		log.Printf("⚠️ Неизвестный admin callback: %s", parts[1])
	}
}

// Заглушка для handleEditProbability
func handleEditProbability(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, techKey string) {
	callback := tgbotapi.NewCallback(callbackQuery.ID, "🎲 Вероятность: пока в разработке")
	bot.Request(callback)
	log.Printf("🛠️ Изменение вероятности для %s от @%s", techKey, callbackQuery.From.UserName)
}

// handleAdminTriggersUICallback - обработка админских триггеров
func handleAdminTriggersUICallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, parts []string, db *sql.DB) {
	if len(parts) < 3 {
		log.Printf("⚠️ Неполный admin triggers callback: %v", parts)
		return
	}

	switch parts[2] {
	case "list":
		// Показать первой страницы админских триггеров
		log.Printf("👑 Админский список триггеров от @%s", callbackQuery.From.UserName)
		showAdminTriggersMenu(bot, callbackQuery, db, 0)
	case "page":
		// Показать конкретную страницу
		if len(parts) < 4 {
			log.Printf("⚠️ Нет номера страницы: %v", parts)
			return
		}
		page, err := strconv.Atoi(parts[3])
		if err != nil {
			log.Printf("❌ Неверный номер страницы: %s", parts[3])
			return
		}
		log.Printf("👑 Админская страница триггеров %d от @%s", page, callbackQuery.From.UserName)
		showAdminTriggersMenu(bot, callbackQuery, db, page)
	default:
		log.Printf("⚠️ Неизвестный admin triggers команда: %s", parts[2])
	}
}

// showAdminMenu показывает админское меню (после нажатия на СВИНОАДМИНКА)
func showAdminMenu(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	text := "🐷 *СвиноАдминка*\n\n" +
		"Выберите действие:"

	// Создаем inline-клавиатуру с четырьмя кнопками ГОРИЗОНТАЛЬНО
	refreshButton := tgbotapi.NewInlineKeyboardButtonData(
		"🔄 Обновить",
		"admin:refresh",
	)
	triggersButton := tgbotapi.NewInlineKeyboardButtonData(
		"📋 Триггеры",
		"admin:triggers:list",
	)
	bdtechButton := tgbotapi.NewInlineKeyboardButtonData(
		"🛠️ БД Тех",
		"admin:bdtech:menu",
	)
	homeButton := tgbotapi.NewInlineKeyboardButtonData(
		"🏠 Главная",
		"admin:home",
	)

	// Четыре кнопки в один ряд (горизонтально)
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(refreshButton, triggersButton, bdtechButton, homeButton),
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
		log.Printf("❌ Ошибка отправки админского меню: %v", err)
	}
}

// handleAdminRefreshTriggers - обновление триггеров из админки
func handleAdminRefreshTriggers(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB) {
	// Проверяем, что это личный чат
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Админский callback из группы, игнорируем: chat_id=%d",
			callbackQuery.Message.Chat.ID)
		return
	}

	// Вызываем существующую логику через виртуальное сообщение
	virtualMsg := &tgbotapi.Message{
		MessageID: callbackQuery.Message.MessageID,
		From:      callbackQuery.From,
		Chat:      callbackQuery.Message.Chat,
		Text:      "/refresh_me",
		Date:      callbackQuery.Message.Date,
	}

	HandleRefreshMeCommand(bot, virtualMsg, db)
	log.Printf("✅ Триггеры обновлены через админку от @%s", callbackQuery.From.UserName)

	// Ждем 3 секунды и возвращаем в стартовое меню
	go func() {
		time.Sleep(3 * time.Second)

		// Проверяем админские права для правильного меню
		if isAdmin(callbackQuery.From.ID) {
			SendAdminMainMenu(bot, callbackQuery.Message.Chat.ID)
		} else {
			SendUserMainMenu(bot, callbackQuery.Message.Chat.ID)
		}

		log.Printf("🔙 Автоматический возврат в стартовое меню для @%s",
			callbackQuery.From.UserName)
	}()
}

// showAdminTriggersMenu показывает админское меню триггеров
func showAdminTriggersMenu(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, page int) {
	// Проверяем, что это личный чат
	if callbackQuery.Message.Chat.Type != "private" {
		log.Printf("⚠️ Админский callback из группы, игнорируем: chat_id=%d",
			callbackQuery.Message.Chat.ID)
		return
	}

	// Генерируем меню страницы с админской навигацией
	menuText, menuKeyboard := GenerateAdminTriggersMenu(page)

	// Редактируем сообщение
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		menuText,
		menuKeyboard,
	)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки админского меню триггеров: %v", err)
	}

	log.Printf("✅ Админское меню триггеров (страница %d) отправлено для @%s",
		page, callbackQuery.From.UserName)
}
