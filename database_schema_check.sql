-- ============================================
-- DATABASE SCHEMA CHECK
-- ============================================

-- Количество записей в таблицах
SELECT 'dict_triggers' as table_name, COUNT(*) as row_count FROM svyno_sobaka_bot.dict_triggers
UNION ALL
SELECT 'triggers', COUNT(*) FROM svyno_sobaka_bot.triggers
UNION ALL
SELECT 'patterns', COUNT(*) FROM svyno_sobaka_bot.patterns
UNION ALL
SELECT 'responses', COUNT(*) FROM svyno_sobaka_bot.responses;

-- Проверка FOREIGN KEY связей
SELECT 
    conname as constraint_name,
    conrelid::regclass || ' → ' || confrelid::regclass as connection
FROM pg_constraint
WHERE contype = 'f' 
  AND conrelid::regclass::text LIKE 'svyno_sobaka_bot.%';

-- Активные триггеры по приоритету (для бота)
SELECT 
    tech_key,
    name,
    priority,
    probability
FROM svyno_sobaka_bot.v_active_triggers 
ORDER BY priority ASC;
