package mybot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendMessageOriginalJSON отправляет оригинальный JSON от Telegram в указанный чат как файл
func SendMessageOriginalJSON(bot *tgbotapi.BotAPI, rawJSON []byte, logChatID int64) {
	// Логируем начало обработки
	log.Printf("📄 Обработка сырого JSON (%d байт)", len(rawJSON))

	// 1. Парсим JSON для извлечения метаданных
	var data map[string]interface{}
	if err := json.Unmarshal(rawJSON, &data); err != nil {
		log.Printf("❌ Ошибка парсинга JSON: %v", err)
		return
	}

	// 2. Генерируем имя файла
	fileName := generateFileName(data)
	log.Printf("📁 Имя файла: %s", fileName)

	// 3. Форматируем JSON с отступами
	formattedJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("❌ Ошибка форматирования JSON: %v", err)
		formattedJSON = rawJSON // используем сырой JSON как fallback
	}

	// 4. Отправляем как файл
	sendJSONAsFile(bot, logChatID, fileName, formattedJSON)
}

// generateFileName генерирует имя файла по правилам: {тип}_{chat_id}_{user_id}_{object_id}.json
func generateFileName(data map[string]interface{}) string {
	var parts []string

	// Определяем тип события
	eventType := getEventType(data)
	parts = append(parts, eventType)

	// Получаем chat_id
	chatID := extractChatID(data)
	parts = append(parts, fmt.Sprintf("%d", chatID))

	// Получаем user_id
	userID := extractUserID(data, eventType)
	parts = append(parts, fmt.Sprintf("%d", userID))

	// Получаем object_id
	objectID := extractObjectID(data, eventType)
	parts = append(parts, objectID)

	// Собираем имя файла
	fileName := strings.Join(parts, "_") + ".json"

	// Ограничиваем длину до 100 символов (с запасом от 255)
	if len(fileName) > 100 {
		// Сохраняем расширение .json
		fileName = fileName[:96] + ".json"
	}

	return fileName
}

// getEventType определяет тип события по структуре JSON
func getEventType(data map[string]interface{}) string {
	// Проверяем все возможные поля Telegram Update
	for _, field := range []string{"message", "edited_message", "channel_post",
		"edited_channel_post", "inline_query", "chosen_inline_result",
		"callback_query", "shipping_query", "pre_checkout_query",
		"poll", "poll_answer", "my_chat_member", "chat_member",
		"chat_join_request"} {
		if _, exists := data[field]; exists {
			return field
		}
	}
	return "unknown"
}

// extractChatID извлекает chat_id из данных
func extractChatID(data map[string]interface{}) int64 {
	// Ищем chat_id в различных структурах
	eventType := getEventType(data)

	switch eventType {
	case "message", "edited_message", "channel_post", "edited_channel_post":
		if msg, ok := data[eventType].(map[string]interface{}); ok {
			if chat, ok := msg["chat"].(map[string]interface{}); ok {
				if id, ok := chat["id"].(float64); ok {
					return int64(id)
				}
			}
		}

	case "callback_query":
		if cb, ok := data["callback_query"].(map[string]interface{}); ok {
			// Если есть сообщение, берем chat_id из него
			if msg, ok := cb["message"].(map[string]interface{}); ok {
				if chat, ok := msg["chat"].(map[string]interface{}); ok {
					if id, ok := chat["id"].(float64); ok {
						return int64(id)
					}
				}
			}
			// Иначе пробуем from
			if from, ok := cb["from"].(map[string]interface{}); ok {
				if id, ok := from["id"].(float64); ok {
					return int64(id) // Это user_id, но если нет chat, используем его
				}
			}
		}

	case "my_chat_member", "chat_member", "chat_join_request":
		if event, ok := data[eventType].(map[string]interface{}); ok {
			if chat, ok := event["chat"].(map[string]interface{}); ok {
				if id, ok := chat["id"].(float64); ok {
					return int64(id)
				}
			}
		}

	case "inline_query", "chosen_inline_result", "shipping_query", "pre_checkout_query":
		if event, ok := data[eventType].(map[string]interface{}); ok {
			if from, ok := event["from"].(map[string]interface{}); ok {
				if id, ok := from["id"].(float64); ok {
					return int64(id) // Это user_id, но если нет chat, используем его
				}
			}
		}
	}

	return 0 // Если chat_id не найден
}

// extractUserID извлекает user_id из данных
func extractUserID(data map[string]interface{}, eventType string) int64 {
	switch eventType {
	case "message", "edited_message", "channel_post", "edited_channel_post":
		if msg, ok := data[eventType].(map[string]interface{}); ok {
			if from, ok := msg["from"].(map[string]interface{}); ok {
				if id, ok := from["id"].(float64); ok {
					return int64(id)
				}
			}
		}

	case "callback_query":
		if cb, ok := data["callback_query"].(map[string]interface{}); ok {
			if from, ok := cb["from"].(map[string]interface{}); ok {
				if id, ok := from["id"].(float64); ok {
					return int64(id)
				}
			}
		}

	case "my_chat_member", "chat_member", "chat_join_request":
		if event, ok := data[eventType].(map[string]interface{}); ok {
			if from, ok := event["from"].(map[string]interface{}); ok {
				if id, ok := from["id"].(float64); ok {
					return int64(id)
				}
			}
		}

	case "inline_query", "chosen_inline_result", "shipping_query", "pre_checkout_query":
		if event, ok := data[eventType].(map[string]interface{}); ok {
			if from, ok := event["from"].(map[string]interface{}); ok {
				if id, ok := from["id"].(float64); ok {
					return int64(id)
				}
			}
		}
	}

	return 0 // Если user_id не найден
}

// extractObjectID извлекает object_id (message_id, callback_id и т.д.)
func extractObjectID(data map[string]interface{}, eventType string) string {
	switch eventType {
	case "message", "edited_message", "channel_post", "edited_channel_post":
		if msg, ok := data[eventType].(map[string]interface{}); ok {
			if msgID, ok := msg["message_id"].(float64); ok {
				return strconv.Itoa(int(msgID))
			}
		}

	case "callback_query":
		if cb, ok := data["callback_query"].(map[string]interface{}); ok {
			if cbID, ok := cb["id"].(string); ok {
				// Обрезаем до 50 символов если нужно
				if len(cbID) > 50 {
					return cbID[:50]
				}
				return cbID
			}
		}

	case "inline_query", "chosen_inline_result":
		if event, ok := data[eventType].(map[string]interface{}); ok {
			if id, ok := event["id"].(string); ok {
				if len(id) > 50 {
					return id[:50]
				}
				return id
			}
		}

	case "shipping_query", "pre_checkout_query":
		if event, ok := data[eventType].(map[string]interface{}); ok {
			if id, ok := event["id"].(string); ok {
				if len(id) > 50 {
					return id[:50]
				}
				return id
			}
		}

	case "poll":
		if poll, ok := data["poll"].(map[string]interface{}); ok {
			if id, ok := poll["id"].(string); ok {
				if len(id) > 50 {
					return id[:50]
				}
				return id
			}
		}

	case "poll_answer":
		if answer, ok := data["poll_answer"].(map[string]interface{}); ok {
			if pollID, ok := answer["poll_id"].(string); ok {
				if len(pollID) > 50 {
					return pollID[:50]
				}
				return pollID
			}
		}
	}

	return "0" // Если object_id не найден
}

// sendJSONAsFile отправляет JSON как файл в Telegram чат
func sendJSONAsFile(bot *tgbotapi.BotAPI, chatID int64, fileName string, jsonData []byte) {
	// Создаем файл для отправки
	file := tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: jsonData,
	}

	// Создаем сообщение с файлом
	msg := tgbotapi.NewDocument(chatID, file)
	msg.Caption = fmt.Sprintf("📄 %s\nРазмер: %.2f KB",
		fileName, float64(len(jsonData))/1024)

	// Отправляем файл
	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Ошибка отправки JSON файла: %v", err)
		// Пробуем отправить как текстовое сообщение (обрезанное)
		if len(jsonData) > 4000 {
			jsonData = jsonData[:4000]
		}
		textMsg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("📄 %s\n```json\n%s\n```",
				fileName, string(jsonData)))
		textMsg.ParseMode = "Markdown"
		bot.Send(textMsg)
	} else {
		log.Printf("✅ JSON файл отправлен: %s (%.2f KB)",
			fileName, float64(len(jsonData))/1024)
	}
}
