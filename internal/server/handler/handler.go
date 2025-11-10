package handler

import (
	"fmt"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/server/interfaces"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	storage interfaces.Storage
	pgx *pgxpool.Pool
}

func New(s interfaces.Storage, pgx *pgxpool.Pool) *Handler {
	return &Handler {
		storage: s,
		pgx: pgx,
	}
}

func (h *Handler) validateMetric(metric *models.Metrics) error {
	// Проверяем тип метрики
	if metric.MType != models.Gauge && metric.MType != models.Counter {
		return fmt.Errorf("invalid metric type: %s", metric.MType)
	}

	// Проверяем наличие ID
	if metric.ID == "" {
		return fmt.Errorf("metric ID is required")
	}

	// Проверяем значения в зависимости от типа
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