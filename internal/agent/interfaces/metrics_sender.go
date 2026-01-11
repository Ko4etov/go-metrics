package interfaces

import "github.com/Ko4etov/go-metrics/internal/models"

// MetricsSender определяет интерфейс отправителя метрик.
type MetricsSender interface {
	SendMetrics(metrics []models.Metrics)
}