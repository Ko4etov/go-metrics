package storage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Ko4etov/go-metrics/internal/models"
)

type MetricsStorage struct {
	metrics map[string]models.Metrics
	mu *sync.Mutex
}

func New() *MetricsStorage {
	return &MetricsStorage{
		metrics: make(map[string]models.Metrics),
		mu: &sync.Mutex{},
	}
}

func (ms *MetricsStorage) Metrics() map[string]models.Metrics {
	return ms.metrics
}

func (ms *MetricsStorage) UpdateMetric(metric models.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
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
	metric, ok := ms.metrics[id]
	return metric, ok
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

func (ms *MetricsStorage) GaugeMetricModel(name string) (*models.Metrics, error) {
	metric, exists := ms.Metric(name)
	if !exists || metric.MType != "gauge" {
		return &models.Metrics{}, fmt.Errorf("gauge metric not found")
	}

	if metric.Value == nil {
		return &models.Metrics{}, fmt.Errorf("invalid gauge value")
	}

	return &metric, nil
}

func (ms *MetricsStorage) CounterMetricModel(name string) (*models.Metrics, error) {
	metric, exists := ms.Metric(name)
	if !exists || metric.MType != "counter" {
		return &models.Metrics{}, fmt.Errorf("counter metric not found")
	}

	if metric.Delta == nil {
		return &models.Metrics{}, fmt.Errorf("invalid counter value")
	}

	return &metric, nil
}

var (
	ErrInvalidType  = errors.New("invalid metric type")
	ErrInvalidValue = errors.New("invalid value for gauge metric")
	ErrInvalidDelta = errors.New("invalid delta for counter metric")
)
