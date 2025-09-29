package repository

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/Ko4etov/go-metrics/internal/models"
)

// Collector собирает метрики из runtime
type Collector struct {
    metrics map[string]models.Metrics
    mu      sync.RWMutex
    pollCount int64
}

// NewCollector создает новый сборщик метрик
func NewCollector() *Collector {
    return &Collector{
        metrics: make(map[string]models.Metrics),
    }
}

// Collect собирает все метрики
func (c *Collector) Collect() {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Собираем метрики из runtime
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)

    // Gauge метрики из runtime
    runtimeMetrics := map[string]float64{
        "Alloc":         float64(stats.Alloc),
        "BuckHashSys":   float64(stats.BuckHashSys),
        "Frees":         float64(stats.Frees),
        "GCCPUFraction": stats.GCCPUFraction,
        "GCSys":         float64(stats.GCSys),
        "HeapAlloc":     float64(stats.HeapAlloc),
        "HeapIdle":      float64(stats.HeapIdle),
        "HeapInuse":     float64(stats.HeapInuse),
        "HeapObjects":   float64(stats.HeapObjects),
        "HeapReleased":  float64(stats.HeapReleased),
        "HeapSys":       float64(stats.HeapSys),
        "LastGC":        float64(stats.LastGC),
        "Lookups":       float64(stats.Lookups),
        "MCacheInuse":   float64(stats.MCacheInuse),
        "MCacheSys":     float64(stats.MCacheSys),
        "MSpanInuse":    float64(stats.MSpanInuse),
        "MSpanSys":      float64(stats.MSpanSys),
        "Mallocs":       float64(stats.Mallocs),
        "NextGC":        float64(stats.NextGC),
        "NumForcedGC":   float64(stats.NumForcedGC),
        "NumGC":         float64(stats.NumGC),
        "OtherSys":      float64(stats.OtherSys),
        "PauseTotalNs":  float64(stats.PauseTotalNs),
        "StackInuse":    float64(stats.StackInuse),
        "StackSys":      float64(stats.StackSys),
        "Sys":           float64(stats.Sys),
        "TotalAlloc":    float64(stats.TotalAlloc),
    }

    for name, value := range runtimeMetrics {
        valueCopy := value
        c.metrics[name] = models.Metrics{
            ID:  name,
            MType:  models.Gauge,
            Value: &valueCopy,
        }
    }

	// PollCount - counter метрика
    c.metrics["PollCount"] = models.Metrics{
        ID:  "PollCount",
        MType:  models.Counter,
        Delta: &c.pollCount,
    }

    var randValue = rand.Float64() * 100

    // RandomValue - gauge метрика со случайным значением
    c.metrics["RandomValue"] = models.Metrics{
        ID:  "RandomValue",
        MType:  models.Gauge,
        Value: &randValue, // Случайное значение от 0 до 100
    }

    // Увеличиваем счетчик опросов
    c.pollCount++
}

// GetMetrics возвращает все собранные метрики
func (c *Collector) GetMetrics() []models.Metrics {
    c.mu.RLock()
    defer c.mu.RUnlock()

    metrics := make([]models.Metrics, 0, len(c.metrics))
    for _, metric := range c.metrics {
        metrics = append(metrics, metric)
    }

    return metrics
}

// GetPollCount возвращает текущее значение счетчика опросов
func (c *Collector) GetPollCount() int64 {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.pollCount
}