// Package interfaces содержит определения интерфейсов для системы сбора метрик.
package interfaces

import "github.com/Ko4etov/go-metrics/internal/models"

// Storage определяет интерфейс хранилища метрик.
type Storage interface {
	GaugeMetric(name string) (string, error)
	CounterMetric(name string) (string, error)
	CounterMetricModel(name string) (*models.Metrics, error)
	GaugeMetricModel(name string) (*models.Metrics, error)
	Metrics() map[string]models.Metrics
	UpdateMetric(metric models.Metrics) error
	Metric(id string) (models.Metrics, bool)
	UpdateMetricsBatch(metrics []models.Metrics) error
	ResetAll()
}