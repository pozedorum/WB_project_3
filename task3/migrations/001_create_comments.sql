-- Базовая таблица комментариев
CREATE TABLE IF NOT EXISTS comments (
    id VARCHAR(36) PRIMARY KEY,
    parent_id VARCHAR(36) NULL REFERENCES comments(id) ON DELETE CASCADE,
    author VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted BOOLEAN DEFAULT FALSE
);

-- Индексы для оптимизации
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id) WHERE deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at) WHERE deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_comments_deleted ON comments(deleted);
CREATE INDEX IF NOT EXISTS idx_comments_content ON comments(content) WHERE deleted = FALSE;

-- Функция и триггер для обновления времени изменения
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE TRIGGER update_comments_updated_at
    BEFORE UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();