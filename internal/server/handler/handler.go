package handler

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/server/interfaces"
)

// Handler обрабатывает HTTP-запросы для работы с метриками.
type Handler struct {
	storage interfaces.Storage // хранилище метрик
	pgx     *pgxpool.Pool      // пул подключений к базе данных
}

// New создает новый обработчик.
func New(s interfaces.Storage, pgx *pgxpool.Pool) *Handler {
	return &Handler{
		storage: s,
		pgx:     pgx,
	}
}

// validateMetric проверяет валидность метрики.
func (h *Handler) validateMetric(metric *models.Metrics) error {
	if metric.MType != models.Gauge && metric.MType != models.Counter {
		return fmt.Errorf("invalid metric type: %s", metric.MType)
	}

	if metric.ID == "" {
		return fmt.Errorf("metric ID is required")
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("value is required for gauge metric")
		}
	case models.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("delta is required for counter metric")
		}
		if *metric.Delta < 0 {
			return fmt.Errorf("delta cannot be negative for counter metric")
		}
	}

	return nil
}