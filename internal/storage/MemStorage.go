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
    // Update обновляет значение метрики
    Update(metric models.Metrics) error
    
    // GetAll возвращает все метрики
    GetAll() map[string]models.Metrics
    
    // UpdateGauge обновляет gauge метрику
    UpdateGauge(name string, value float64) error
    
    // UpdateCounter обновляет counter метрику (добавляет значение)
    UpdateCounter(name string, delta int64) error
}

func GetInstance() *MemStorage {
    once.Do(func() {
        instance = &MemStorage{
            metrics: make(map[string]models.Metrics),
        }
    })
    return instance
}

// GetAllMetrics возвращает все метрики
func (ms *MemStorage) GetAllMetrics() map[string]models.Metrics {
    return ms.metrics
}

// UpdateMetric обновляет или создает метрику
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

var (
    ErrInvalidType  = errors.New("invalid metric type")
    ErrInvalidValue = errors.New("invalid value for gauge metric")
    ErrInvalidDelta = errors.New("invalid delta for counter metric")
)