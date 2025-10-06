package collector

import "github.com/Ko4etov/go-metrics/internal/models"

type MetricsCollector interface {
	Collect()
	Metrics() []models.Metrics
	PollCount() int64
}
