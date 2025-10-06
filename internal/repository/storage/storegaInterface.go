package storage

import "github.com/Ko4etov/go-metrics/internal/models"

type Storage interface {
	GaugeMetric(name string) (string, error)
	CounterMetric(name string) (string, error)
	Metrics() map[string]models.Metrics
	UpdateMetric(metric models.Metrics) error
	Metric(id string) (models.Metrics, bool)
	ResetAll()
}
