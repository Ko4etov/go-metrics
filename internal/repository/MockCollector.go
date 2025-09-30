package repository

import (
	"github.com/Ko4etov/go-metrics/internal/models"
)

type MockCollector struct {
    CollectCount int
}

func (m *MockCollector) Collect() {
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