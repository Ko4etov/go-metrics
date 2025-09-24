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
func (ms *MemStorage) GetAllMetrics() map[string]interface{} {
    
    result := make(map[string]interface{})
    for k, v := range ms.metrics {
        result[k] = v
    }
    return result
}

// UpdateMetric обновляет или создает метрику
func (ms *MemStorage) UpdateMetric(metric models.Metrics) error {

    switch metric.MType {
    case "gauge":
        if metric.Value == nil {
            return ErrInvalidValue
        }
        ms.metrics[metric.ID] = metric
    
    case "counter":
        if metric.Delta == nil {
            return ErrInvalidDelta
        }
        
        // Для counter добавляем значение к существующему
        if existing, exists := ms.metrics[metric.ID]; exists && existing.MType == "counter" {
            newDelta := *existing.Delta + *metric.Delta
            metric.Delta = &newDelta
        }
        ms.metrics[metric.ID] = metric
    
    default:
        return ErrInvalidType
    }
    
    return nil
}

var (
    ErrInvalidType  = errors.New("invalid metric type")
    ErrInvalidValue = errors.New("invalid value for gauge metric")
    ErrInvalidDelta = errors.New("invalid delta for counter metric")
)