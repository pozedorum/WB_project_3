-- migrations/001_init_tables.up.sql
CREATE TABLE items (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price BIGINT NOT NULL CHECK (price > 0),
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE item_history (
    id BIGSERIAL PRIMARY KEY,
    item_id BIGINT REFERENCES items(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL CHECK (action IN ('CREATE', 'UPDATE', 'DELETE')),
    changed_by VARCHAR(100) NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Триггерная функция для логирования изменений
CREATE OR REPLACE FUNCTION log_item_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO item_history (item_id, action, changed_by)
        VALUES (NEW.id, 'CREATE', NEW.created_by);
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO item_history (item_id, action, changed_by)
        VALUES (NEW.id, 'UPDATE', NEW.created_by);
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO item_history (item_id, action, changed_by)
        VALUES (OLD.id, 'DELETE', OLD.created_by);
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Триггеры на все операции
CREATE TRIGGER items_history_trigger
    AFTER INSERT OR UPDATE OR DELETE ON items
    FOR EACH ROW
    EXECUTE FUNCTION log_item_changes();

-- Индексы для производительности
CREATE INDEX idx_item_history_item_id ON item_history(item_id);
CREATE INDEX idx_item_history_changed_at ON item_history(changed_at);
CREATE INDEX idx_item_history_changed_by ON item_history(changed_by);