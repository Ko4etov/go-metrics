package interfaces

import "github.com/Ko4etov/go-metrics/internal/models"

type MetricsSender interface {
	SendMetrics(metrics []models.Metrics)
}
