CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    category VARCHAR(100) NOT NULL,
    description TEXT,
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого поиска и аналитики
CREATE INDEX idx_sales_date ON sales(date);
CREATE INDEX idx_sales_category ON sales(category);
CREATE INDEX idx_sales_type ON sales(type);
CREATE INDEX idx_sales_date_category ON sales(date, category);