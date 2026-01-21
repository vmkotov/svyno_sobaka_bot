// ============================================================================
// ФАЙЛ: admin.go
// Ядро админ-системы: проверка прав администратора
// ============================================================================
package mybot

// adminIDs - хардкод ID администраторов
// Только эти пользователи имеют доступ к админским функциям
var adminIDs = map[int64]bool{
    266468924: true, // Первый администратор
    435014351: true, // Второй администратор
}

// isAdmin проверяет, является ли пользователь администратором
// Используется при обработке команд и callback-запросов
func isAdmin(userID int64) bool {
    // Простая проверка по карте ID
    return adminIDs[userID]
}

// checkAdminAccess проверяет доступ к админским функциям
// Используется для callback-запросов с префиксом "admin:"
func checkAdminAccess(userID int64, callbackData string) bool {
    // Если callback начинается с "admin:" - проверяем права
    if len(callbackData) >= 6 && callbackData[:6] == "admin:" {
        return isAdmin(userID)
    }
    // Для не-админских callback доступ открыт
    return true
}
