package repository

import (
	"sync"

	"github.com/Ko4etov/go-metrics/internal/models"
)

type MockCollector struct {
    CollectCount int
    mu           sync.Mutex
}

func (m *MockCollector) Collect() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.CollectCount++
}

func (m *MockCollector) GetMetrics() []models.Metrics {
    value := 1.0
    return []models.Metrics{
        {ID: "TestMetric", MType: models.Gauge, Value: &value},
    }
}

func (m *MockCollector) GetPollCount() int64 {
    return int64(m.CollectCount)
}