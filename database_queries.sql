-- ============================================
-- DATABASE QUERIES FOR SVINO_SOBAKA_BOT
-- Created: четверг, 15 января 2026 г. 16:47:34 (MSK)
-- ============================================

-- 1. Мастер-справочник триггеров
SELECT * FROM svyno_sobaka_bot.dict_triggers ORDER BY id;

-- 2. Версии настроек триггеров
SELECT * FROM svyno_sobaka_bot.triggers ORDER BY priority;

-- 3. Паттерны поиска
SELECT * FROM svyno_sobaka_bot.patterns ORDER BY tech_key, type;

-- 4. Ответы на триггеры
SELECT * FROM svyno_sobaka_bot.responses ORDER BY tech_key, weight DESC;

-- 5. Активные триггеры (dt_end в будущем)
SELECT * FROM svyno_sobaka_bot.v_active_triggers;

-- 6. Активные паттерны
SELECT * FROM svyno_sobaka_bot.v_active_patterns;

-- 7. Активные ответы
SELECT * FROM svyno_sobaka_bot.v_active_responses;
