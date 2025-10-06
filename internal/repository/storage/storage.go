package storage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Ko4etov/go-metrics/internal/models"
)

var (
	instance Storage
	once     sync.Once
)

type MetricsStorage struct {
	metrics map[string]models.Metrics
}

func New() Storage {
	once.Do(func() {
		instance = &MetricsStorage{
			metrics: make(map[string]models.Metrics),
		}
	})
	return instance
}

func (ms *MetricsStorage) Metrics() map[string]models.Metrics {
	return ms.metrics
}

func (ms *MetricsStorage) UpdateMetric(metric models.Metrics) error {
	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return ErrInvalidValue
		}
		ms.metrics[metric.ID] = metric

	case models.Counter:
		if metric.Delta == nil {
			return ErrInvalidDelta
		}

		// Для counter добавляем значение к существующему
		if existing, exists := ms.metrics[metric.ID]; exists && existing.MType == models.Counter {
			newDelta := *existing.Delta + *metric.Delta
			metric.Delta = &newDelta
		}
		ms.metrics[metric.ID] = metric

	default:
		return ErrInvalidType
	}

	return nil
}

func (ms *MetricsStorage) Metric(id string) (models.Metrics, bool) {
	metric, exists := ms.metrics[id]
	return metric, exists
}

func (ms *MetricsStorage) ResetAll() {
	ms.metrics = make(map[string]models.Metrics)
}

func (ms *MetricsStorage) GaugeMetric(name string) (string, error) {
	metric, exists := ms.Metric(name)
	if !exists || metric.MType != "gauge" {
		return "", fmt.Errorf("gauge metric not found")
	}

	if metric.Value == nil {
		return "", fmt.Errorf("invalid gauge value")
	}

	return fmt.Sprintf("%g", *metric.Value), nil
}

func (ms *MetricsStorage) CounterMetric(name string) (string, error) {
	metric, exists := ms.Metric(name)
	if !exists || metric.MType != "counter" {
		return "", fmt.Errorf("counter metric not found")
	}

	if metric.Delta == nil {
		return "", fmt.Errorf("invalid counter value")
	}

	return fmt.Sprintf("%d", *metric.Delta), nil
}

var (
	ErrInvalidType  = errors.New("invalid metric type")
	ErrInvalidValue = errors.New("invalid value for gauge metric")
	ErrInvalidDelta = errors.New("invalid delta for counter metric")
)
