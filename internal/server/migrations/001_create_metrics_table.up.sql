-- Создание таблицы для хранения метрик
CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('gauge', 'counter')),
    delta BIGINT,
    value DOUBLE PRECISION,
    hash VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id, type)
);

-- Индекс для быстрого поиска по типу метрики
CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type);

-- Индекс для быстрого поиска по хешу
CREATE INDEX IF NOT EXISTS idx_metrics_hash ON metrics(hash);

-- Индекс для временных меток
CREATE INDEX IF NOT EXISTS idx_metrics_updated_at ON metrics(updated_at);

-- Комментарии к таблице и колонкам
COMMENT ON TABLE metrics IS 'Таблица для хранения метрик приложения';
COMMENT ON COLUMN metrics.id IS 'Уникальный идентификатор метрики';
COMMENT ON COLUMN metrics.type IS 'Тип метрики: gauge или counter';
COMMENT ON COLUMN metrics.delta IS 'Значение для counter метрик';
COMMENT ON COLUMN metrics.value IS 'Значение для gauge метрик';
COMMENT ON COLUMN metrics.hash IS 'Хеш для проверки целостности данных';
COMMENT ON COLUMN metrics.created_at IS 'Время создания записи';
COMMENT ON COLUMN metrics.updated_at IS 'Время последнего обновления записи';