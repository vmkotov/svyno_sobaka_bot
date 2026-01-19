-- ============================================
-- DATABASE QUERIES FOR SVINO_SOBAKA_BOT
-- Created: четверг, 15 января 2026 г. 16:47:34 (MSK)
-- Updated: понедельник, 19 января 2026 г. (полная схема)
-- ============================================

-- ======================
-- 📊 СХЕМА MAIN (логи)
-- ======================

-- 1. Быстрые логи сообщений
SELECT * FROM main.messages_log ORDER BY created_at DESC LIMIT 100;

-- 2. Статистика бота
SELECT * FROM main.bot_stats ORDER BY created_at DESC LIMIT 50;

-- 3. Парсинг сообщений по словам
SELECT * FROM main.message_words ORDER BY word_date DESC LIMIT 100;

-- 4. Распарсенные сообщения
SELECT * FROM main.parsed_messages ORDER BY parsed_at DESC LIMIT 100;

-- 5. Репозиторий SQL скриптов
SELECT script_name, created_at FROM main.sql_scripts_repository ORDER BY created_at DESC LIMIT 20;

-- ======================
-- ⚙️ СИСТЕМА ТРИГГЕРОВ
-- ======================

-- 6. Мастер-справочник триггеров
SELECT * FROM svyno_sobaka_bot.dict_triggers ORDER BY id;

-- 7. Версии настроек триггеров
SELECT * FROM svyno_sobaka_bot.triggers ORDER BY priority;

-- 8. Паттерны поиска
SELECT * FROM svyno_sobaka_bot.patterns ORDER BY tech_key, type;

-- 9. Ответы на триггеры
SELECT * FROM svyno_sobaka_bot.responses ORDER BY tech_key, weight DESC;

-- ======================
-- 👥 ПОЛЬЗОВАТЕЛИ И ЧАТЫ
-- ======================

-- 10. Пользователи Telegram
SELECT * FROM svyno_sobaka_bot.users ORDER BY user_id;

-- 11. Чаты Telegram
SELECT * FROM svyno_sobaka_bot.chats ORDER BY chat_id;

-- ======================
-- 💬 СООБЩЕНИЯ И МЕДИА
-- ======================

-- 12. Сообщения (ядро)
SELECT * FROM svyno_sobaka_bot.messages ORDER BY message_date DESC LIMIT 100;

-- 13. Медиафайлы сообщений
SELECT * FROM svyno_sobaka_bot.message_media ORDER BY media_id DESC LIMIT 100;

-- 14. Связи между сообщениями (reply/forward/edit)
SELECT * FROM svyno_sobaka_bot.message_references ORDER BY chat_id, message_id;

-- 15. Упоминания пользователей
SELECT * FROM svyno_sobaka_bot.message_mentions ORDER BY mention_id DESC LIMIT 100;

-- ======================
-- 🐷 СВИНОСОБАКА СИСТЕМА
-- ======================

-- 16. Кандидаты в свинособаки
SELECT * FROM svyno_sobaka_bot.svyno_sobaka_candidates ORDER BY chat_id;

-- 17. Свинособаки дня (рассылка)
SELECT * FROM svyno_sobaka_bot.svyno_sobaka_of_the_day WHERE dt_date_only = CURRENT_DATE ORDER BY chat_id;

-- ======================
-- 📝 ЛОГИРОВАНИЕ И МОНИТОРИНГ
-- ======================

-- 18. Логи процедур (с типом ERROR/LOG)
SELECT * FROM svyno_sobaka_bot.procedure_logs ORDER BY n_round_id DESC, v_record_type LIMIT 100;

-- ======================
-- 🔍 ВЬЮХИ (v_*) - АКТИВНЫЕ ДАННЫЕ
-- ======================

-- 19. Активные триггеры (dt_end в будущем)
SELECT * FROM svyno_sobaka_bot.v_active_triggers;

-- 20. Активные паттерны
SELECT * FROM svyno_sobaka_bot.v_active_patterns;

-- 21. Активные ответы
SELECT * FROM svyno_sobaka_bot.v_active_responses;

-- ======================
-- 📈 АНАЛИТИЧЕСКИЕ ЗАПРОСЫ
-- ======================

-- 22. Статистика по типам медиа
SELECT media_type, COUNT(*) as count, AVG(file_size) as avg_size FROM svyno_sobaka_bot.message_media GROUP BY media_type ORDER BY count DESC;

-- 23. Активность по часам (последние 7 дней)
SELECT EXTRACT(HOUR FROM message_date) as hour, COUNT(*) as message_count FROM svyno_sobaka_bot.messages WHERE message_date > NOW() - INTERVAL '7 days' GROUP BY hour ORDER BY hour;

-- 24. ТОП-10 активных пользователей
SELECT u.user_id, u.username, u.first_name, COUNT(m.message_id) as message_count FROM svyno_sobaka_bot.users u JOIN svyno_sobaka_bot.messages m ON u.user_id = m.user_id GROUP BY u.user_id, u.username, u.first_name ORDER BY message_count DESC LIMIT 10;

-- 25. ТОП-10 активных чатов
SELECT c.chat_id, c.title, c.username, COUNT(m.message_id) as message_count FROM svyno_sobaka_bot.chats c JOIN svyno_sobaka_bot.messages m ON c.chat_id = m.chat_id GROUP BY c.chat_id, c.title, c.username ORDER BY message_count DESC LIMIT 10;

-- 26. Самые частые паттерны триггеров
SELECT p.pattern_text, p.type, COUNT(*) as usage_count FROM svyno_sobaka_bot.patterns p GROUP BY p.pattern_text, p.type ORDER BY usage_count DESC LIMIT 20;

-- 27. Логи процедур с ошибками (только ERROR)
SELECT * FROM svyno_sobaka_bot.procedure_logs WHERE v_record_type = 'ERROR' ORDER BY created_at DESC LIMIT 50;

-- ======================
-- 🗄️ СТРУКТУРА БАЗЫ ДАННЫХ
-- ======================

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
    WHERE t.table_schema IN ('public', 'svyno_sobaka_bot', 'main')
      AND t.table_type = 'BASE TABLE'
)

SELECT jsonb_pretty(
    jsonb_build_object(
        'database', current_database(),
        'generated_at', now(),
        'tables', jsonb_agg(table_data ORDER BY table_schema, table_name)
    )
) as database_structure;