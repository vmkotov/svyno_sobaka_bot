package mybot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

// =============================================
// СТРУКТУРЫ ДАННЫХ ДЛЯ JSON КОНФИГА
// =============================================

type Pattern struct {
	PatternID   int    `json:"pattern_id"`
	PatternText string `json:"pattern_text"`
	PatternType string `json:"pattern_type"`
}

type Response struct {
	ResponseID     int    `json:"response_id"`
	ResponseText   string `json:"response_text"`
	ResponseWeight int    `json:"response_weight"`
}

type Trigger struct {
	TriggerID   int        `json:"trigger_id"`
	TriggerName string     `json:"trigger_name"`
	TechKey     string     `json:"tech_key"`
	Priority    int        `json:"priority"`
	Probability float64    `json:"probability"`
	Patterns    []Pattern  `json:"patterns"`
	Responses   []Response `json:"responses"`
}

type TriggerConfig []Trigger

// =============================================
// ГЛОБАЛЬНЫЕ ПЕРЕМЕННЫЕ
// =============================================

var (
	triggerConfig TriggerConfig
	configMutex   sync.RWMutex
	randSource    = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// =============================================
// ОСНОВНЫЕ ФУНКЦИИ
// =============================================

// LoadTriggerConfig загружает конфигурацию триггеров из БД
func LoadTriggerConfig(db *sql.DB) error {
	log.Printf("🗃️ Загрузка конфигурации триггеров из БД...")

	if db == nil {
		log.Printf("📭 БД не подключена, триггеры отключены")
		return fmt.Errorf("БД не подключена")
	}

	// Получаем JSON из БД
	jsonData, err := GetTriggersConfigJSON(db)
	if err != nil {
		log.Printf("❌ Ошибка загрузки из БД: %v", err)
		return err
	}

	// Парсим JSON
	var config TriggerConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		log.Printf("❌ Ошибка парсинга JSON из БД: %v", err)
		return err
	}

	// Сортируем триггеры по приоритету
	sort.Slice(config, func(i, j int) bool {
		return config[i].Priority < config[j].Priority
	})

	// Сохраняем в глобальную переменную
	configMutex.Lock()
	triggerConfig = config
	configMutex.Unlock()

	log.Printf("✅ Загружено %d триггеров из БД", len(config))

	// Выводим информацию о загруженных триггерах
	for i, trigger := range config {
		log.Printf("   %2d. %-30s (приоритет: %2d, вероятность: %.0f%%, ответов: %d)",
			i+1, trigger.TriggerName, trigger.Priority,
			trigger.Probability*100, len(trigger.Responses))
	}

	return nil
}

// GetTriggerConfig возвращает конфигурацию (потокобезопасно)
func GetTriggerConfig() TriggerConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return triggerConfig
}

// normalizeText приводит текст к нижнему регистру и удаляет знаки препинания
// (как в оригинальных триггерных модулях)
func normalizeText(text string) string {
	// 1. К нижнему регистру
	text = strings.ToLower(text)

	// 2. Удаляем знаки препинания: ,.!?- (и множественные пробелы)
	replacer := strings.NewReplacer(
		",", " ",
		".", " ",
		"!", " ",
		"?", " ",
		"-", " ",
		"  ", " ", // двойные пробелы -> одинарные
	)

	text = replacer.Replace(text)

	// 3. Убираем лишние пробелы
	text = strings.TrimSpace(text)

	return text
}

// GetTriggerByTechKey возвращает триггер по техническому ключу
func GetTriggerByTechKey(techKey string) *Trigger {
	configMutex.RLock()
	defer configMutex.RUnlock()

	for _, trigger := range triggerConfig {
		if trigger.TechKey == techKey {
			// Возвращаем копию, чтобы избежать гонок данных
			return &Trigger{
				TriggerID:   trigger.TriggerID,
				TriggerName: trigger.TriggerName,
				TechKey:     trigger.TechKey,
				Priority:    trigger.Priority,
				Probability: trigger.Probability,
				Patterns:    append([]Pattern{}, trigger.Patterns...),
				Responses:   append([]Response{}, trigger.Responses...),
			}
		}
	}
	return nil
}

// Log - экспортируем логгер для UI
