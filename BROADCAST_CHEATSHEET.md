# 📢 Шпаргалка: Управление рассылкой свинособаки дня

## 1. 🚀 ЗАПУСТИТЬ ВРУЧНУЮ
\`\`\`bash
curl -X POST https://bba74t16lphcg8vfa4o3.containers.yandexcloud.net/admin/broadcast \\
  -H "X-Broadcast-Secret: change-me-in-production"
\`\`\`

## 2. 📅 ПОСТАВИТЬ НА РАСПИСАНИЕ
\`\`\`bash
# Создать триггер (16:00 МСК = 13:00 UTC)
yc serverless trigger create timer \\
  --name svyno-daily-1600 \\
  --cron-expression "00 13 * * ? *" \\
  --invoke-container-id bba74t16lphcg8vfa4o3 \\
  --invoke-container-path "/admin/broadcast" \\
  --invoke-container-service-account-id aje0eno6g4o1o94901fu

# Удалить триггер
yc serverless trigger delete <ID_триггера>

# Список триггеров
yc serverless trigger list
\`\`\`

**Формат cron:** \`"минуты час * * ? *"\` (UTC время)  
**Текущий триггер:** \`svyno-daily-1600\` ✅ Активен
