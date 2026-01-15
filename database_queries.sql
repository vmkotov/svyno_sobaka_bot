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

-- Полная структура базы данных в формате JSON
-- Генерация: четверг, 15 января 2026 г. 19:41:31 (MSK)

-- Полная структура базы данных в формате JSON
WITH table_info AS (
    SELECT 
        t.table_schema,
        t.table_name,
        jsonb_build_object(
            'schema', t.table_schema,
            'table', t.table_name,
            'type', t.table_type,
            'columns', (
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'name', c.column_name,
                        'type', c.data_type,
                        'nullable', c.is_nullable,
                        'default', c.column_default
                    )
                    ORDER BY c.ordinal_position
                )
                FROM information_schema.columns c
                WHERE c.table_schema = t.table_schema 
                  AND c.table_name = t.table_name
            ),
            'indexes', (
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'name', i.indexname,
                        'definition', i.indexdef
                    )
                )
                FROM pg_indexes i
                WHERE i.schemaname = t.table_schema 
                  AND i.tablename = t.table_name
            ),
            'constraints', (
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'name', tc.constraint_name,
                        'type', tc.constraint_type,
                        'definition', pg_get_constraintdef(c.oid)
                    )
                )
                FROM information_schema.table_constraints tc
                JOIN pg_constraint c ON c.conname = tc.constraint_name
                WHERE tc.table_schema = t.table_schema 
                  AND tc.table_name = t.table_name
            )
        ) as table_data
    FROM information_schema.tables t
    WHERE t.table_schema IN ('public', 'svyno_sobaka_bot')
      AND t.table_type = 'BASE TABLE'
)

SELECT jsonb_pretty(
    jsonb_build_object(
        'database', current_database(),
        'generated_at', now(),
        'tables', jsonb_agg(table_data ORDER BY table_schema, table_name)
    )
) as database_structure;
