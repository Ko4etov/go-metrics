package storage

import (
	"errors"
	"sync"

	"github.com/Ko4etov/go-metrics/internal/models"
)

var (
	instance *MemStorage
	once     sync.Once
)

type MemStorage struct {
	metrics map[string]models.Metrics
}

type Storage interface {
	GetAllMetrics() map[string]models.Metrics

    UpdateMetric(metric models.Metrics) error

    GetMetric(id string) (models.Metrics, bool)

    ResetAll()
}

func GetInstance() *MemStorage {
	once.Do(func() {
		instance = &MemStorage{
			metrics: make(map[string]models.Metrics),
		}
	})
	return instance
}

func (ms *MemStorage) GetAllMetrics() map[string]models.Metrics {
	return ms.metrics
}

func (ms *MemStorage) UpdateMetric(metric models.Metrics) error {
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

func (ms *MemStorage) GetMetric(id string) (models.Metrics, bool) {
	metric, exists := ms.metrics[id]
	return metric, exists
}

func (ms *MemStorage) ResetAll() {
    ms.metrics = make(map[string]models.Metrics)
}

var (
	ErrInvalidType  = errors.New("invalid metric type")
	ErrInvalidValue = errors.New("invalid value for gauge metric")
	ErrInvalidDelta = errors.New("invalid delta for counter metric")
)
