package mybot

import (
    "fmt"
    "strings"
    "time"
)

// ================= ОБЩИЕ УТИЛИТЫ =================

// FormatDuration - форматирование времени
func FormatDuration(d time.Duration) string {
    if d < time.Second {
        return fmt.Sprintf("%dms", d.Milliseconds())
    }
    if d < time.Minute {
        return fmt.Sprintf("%.1fs", d.Seconds())
    }
    return fmt.Sprintf("%.1fmin", d.Minutes())
}

// TruncateString - обрезание строки
func TruncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}

// ContainsAny - проверка содержит ли строка любую из подстрок
func ContainsAny(s string, substrings []string) bool {
    for _, sub := range substrings {
        if strings.Contains(s, sub) {
            return true
        }
    }
    return false
}

// UniqueStrings - удаление дубликатов из слайса строк
func UniqueStrings(items []string) []string {
    seen := make(map[string]bool)
    result := []string{}
    
    for _, item := range items {
        if !seen[item] {
            seen[item] = true
            result = append(result, item)
        }
    }
    
    return result
}

// ParseBool - безопасный парсинг булевых значений
func ParseBool(s string) bool {
    switch strings.ToLower(s) {
    case "true", "1", "yes", "y", "on":
        return true
    default:
        return false
    }
}
