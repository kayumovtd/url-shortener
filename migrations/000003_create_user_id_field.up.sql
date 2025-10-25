-- Можно было бы сделать таблицу пользователей и ссылаться на нее через внешний ключ,
-- но для простоты задачи ограничимся просто полем user_id в таблице urls.
ALTER TABLE urls ADD COLUMN IF NOT EXISTS user_id TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
