package mybot

import (
    "strings"
)

// escapeMarkdown экранирует символы Markdown (общая функция)
func escapeMarkdown(text string) string {
    specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
    result := text
    for _, char := range specialChars {
        result = strings.ReplaceAll(result, char, "\\"+char)
    }
    return result
}
