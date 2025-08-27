
CREATE TABLE IF NOT EXISTS short_urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    clicks_count INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS url_clicks (
    id SERIAL PRIMARY KEY,
    short_url_id INTEGER NOT NULL REFERENCES short_urls(id) ON DELETE CASCADE,
    user_agent TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_short_urls_short_code ON short_urls(short_code);
CREATE INDEX IF NOT EXISTS idx_url_clicks_short_url_id ON url_clicks(short_url_id);
CREATE INDEX IF NOT EXISTS idx_url_clicks_created_at ON url_clicks(created_at);