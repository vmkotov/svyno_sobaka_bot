ПЛАН СОЗДАНИЯ СХЕМЫ БАЗЫ ДАННЫХ ДЛЯ TELEGRAM СООБЩЕНИЙ

ОБЗОР
Создаем нормализованную схему (3НФ) для хранения Telegram сообщений.
5 таблиц + 5 процедур для удобной работы.

ТАБЛИЦЫ

1. ТАБЛИЦА users - ПОЛЬЗОВАТЕЛИ TELEGRAM

CREATE TABLE users (
    user_id BIGINT PRIMARY KEY,
    is_bot BOOLEAN NOT NULL DEFAULT false,
    first_name TEXT,
    last_name TEXT,
    username TEXT,
    language_code VARCHAR(10),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    INDEX idx_users_username (username),
    INDEX idx_users_updated (updated_at)
);

Назначение: Справочник всех пользователей (отправители, упоминания).

2. ТАБЛИЦА chats - ЧАТЫ (БЕСЕДЫ, ГРУППЫ, КАНАЛЫ)

CREATE TABLE chats (
    chat_id BIGINT PRIMARY KEY,
    chat_type VARCHAR(20) NOT NULL,
    title TEXT,
    username TEXT,
    first_name TEXT,
    last_name TEXT,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    INDEX idx_chats_type (chat_type),
    INDEX idx_chats_username (username),
    INDEX idx_chats_updated (updated_at)
);

Назначение: Справочник всех чатов.
chat_type значения: 'private', 'group', 'supergroup', 'channel'

3. ТАБЛИЦА messages - ОСНОВНАЯ ТАБЛИЦА СООБЩЕНИЙ

CREATE TABLE messages (
    chat_id BIGINT NOT NULL,
    message_id INT NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    
    text TEXT,
    caption TEXT,
    
    user_id BIGINT NOT NULL,
    is_bot BOOLEAN NOT NULL DEFAULT false,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    PRIMARY KEY (chat_id, message_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (chat_id) REFERENCES chats(chat_id),
    
    INDEX idx_messages_date (date),
    INDEX idx_messages_user (user_id),
    INDEX idx_messages_created (created_at)
);

Назначение: Ядро системы, хранение всех сообщений.

4. ТАБЛИЦА message_media - МЕДИАФАЙЛЫ СООБЩЕНИЙ

CREATE TABLE message_media (
    media_id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    message_id INT NOT NULL,
    
    media_type VARCHAR(20) NOT NULL,
    
    file_id VARCHAR(255) NOT NULL,
    file_unique_id VARCHAR(255) NOT NULL,
    file_size INT,
    
    width INT,
    height INT,
    duration INT,
    mime_type VARCHAR(100),
    file_name TEXT,
    emoji VARCHAR(10),
    
    FOREIGN KEY (chat_id, message_id) REFERENCES messages(chat_id, message_id),
    
    INDEX idx_media_type (media_type),
    INDEX idx_media_file_id (file_id),
    INDEX idx_media_message (chat_id, message_id)
);

Назначение: Хранение всех медиафайлов.
media_type значения: 'photo', 'document', 'sticker', 'video', 'voice', 'audio'

5. ТАБЛИЦА message_references - СВЯЗИ МЕЖДУ СООБЩЕНИЯМИ

CREATE TABLE message_references (
    chat_id BIGINT NOT NULL,
    message_id INT NOT NULL,
    ref_type VARCHAR(20) NOT NULL,
    
    ref_chat_id BIGINT,
    ref_message_id INT,
    ref_user_id BIGINT,
    
    PRIMARY KEY (chat_id, message_id, ref_type),
    FOREIGN KEY (chat_id, message_id) REFERENCES messages(chat_id, message_id),
    FOREIGN KEY (ref_user_id) REFERENCES users(user_id),
    
    INDEX idx_ref_type (ref_type),
    INDEX idx_ref_target (ref_chat_id, ref_message_id)
);

Назначение: Хранение связей между сообщениями.
ref_type значения: 'reply', 'forward', 'edit'

ПРОЦЕДУРЫ

1. ПРОЦЕДУРА upsert_user - ОБНОВЛЕНИЕ/ДОБАВЛЕНИЕ ПОЛЬЗОВАТЕЛЯ

CREATE PROCEDURE upsert_user(
    p_user_id BIGINT,
    p_is_bot BOOLEAN,
    p_first_name TEXT,
    p_last_name TEXT,
    p_username TEXT,
    p_language_code TEXT
)

Параметры: 6
Назначение: Добавить или обновить пользователя

2. ПРОЦЕДУРА upsert_chat - ОБНОВЛЕНИЕ/ДОБАВЛЕНИЕ ЧАТА

CREATE PROCEDURE upsert_chat(
    p_chat_id BIGINT,
    p_chat_type TEXT,
    p_title TEXT,
    p_username TEXT,
    p_first_name TEXT,
    p_last_name TEXT
)

Параметры: 6
Назначение: Добавить или обновить чат

3. ПРОЦЕДУРА insert_message - ДОБАВЛЕНИЕ СООБЩЕНИЯ

CREATE PROCEDURE insert_message(
    p_chat_id BIGINT,
    p_message_id INT,
    p_date TIMESTAMPTZ,
    p_text TEXT,
    p_caption TEXT,
    p_user_id BIGINT,
    p_is_bot BOOLEAN
)

Параметры: 7
Назначение: Добавить основное сообщение

4. ПРОЦЕДУРА insert_message_media - ДОБАВЛЕНИЕ МЕДИА

CREATE PROCEDURE insert_message_media(
    p_chat_id BIGINT,
    p_message_id INT,
    p_media_type TEXT,
    p_file_id TEXT,
    p_file_unique_id TEXT,
    p_width INT,
    p_height INT,
    p_duration INT,
    p_mime_type TEXT,
    p_file_name TEXT,
    p_emoji TEXT
)

Параметры: 11
Назначение: Добавить медиафайл к сообщению

5. ПРОЦЕДУРА insert_message_reference - ДОБАВЛЕНИЕ СВЯЗИ

CREATE PROCEDURE insert_message_reference(
    p_chat_id BIGINT,
    p_message_id INT,
    p_ref_type TEXT,
    p_ref_chat_id BIGINT,
    p_ref_message_id INT,
    p_ref_user_id BIGINT
)

Параметры: 6
Назначение: Добавить связь между сообщениями

ПОРЯДОК СОЗДАНИЯ

1. Создать таблицы в порядке:
   - users
   - chats  
   - messages
   - message_media
   - message_references

2. Создать процедуры в любом порядке

3. Протестировать на примере:
   - Создать тестового пользователя и чат
   - Добавить сообщение
   - Добавить медиа (если нужно)
   - Добавить ссылку (если нужно)

ПРИМЕЧАНИЯ

- Все таблицы в схеме main (если не указано иное)
- Использовать ON CONFLICT для upsert операций
- Добавить триггеры для автоматического обновления updated_at
- Рассмотреть партиционирование messages по дате при больших объемах

КОНЕЦ ДОКУМЕНТА
