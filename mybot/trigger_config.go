package mybot

import (
    "encoding/json"
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
    TriggerID    int        `json:"trigger_id"`
    TriggerName  string     `json:"trigger_name"`
    TechKey      string     `json:"tech_key"`
    Priority     int        `json:"priority"`
    Probability  float64    `json:"probability"`
    Patterns     []Pattern  `json:"patterns"`
    Responses    []Response `json:"responses"`
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

// LoadTriggerConfig загружает конфигурацию из встроенной строки JSON
func LoadTriggerConfig() error {
    log.Printf("📁 Загрузка встроенной конфигурации триггеров")
    
    var config TriggerConfig
    if err := json.Unmarshal([]byte(TriggersJSON), &config); err != nil {
        log.Printf("❌ Ошибка парсинга JSON: %v", err)
        return err
    }
    
    // Сортируем триггеры по приоритету
    sort.Slice(config, func(i, j int) bool {
        return config[i].Priority < config[j].Priority
    })
    
    configMutex.Lock()
    triggerConfig = config
    configMutex.Unlock()
    
    log.Printf("✅ Загружено %d триггеров", len(config))
    
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
