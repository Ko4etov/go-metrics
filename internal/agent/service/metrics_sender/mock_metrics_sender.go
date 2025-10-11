package metrics_sender

import (
	"sync"

	"github.com/Ko4etov/go-metrics/internal/models"
)

type MockMetricsSenderService struct {
    SendCount int
    mu        sync.Mutex
}

func (m *MockMetricsSenderService) SendMetrics(metrics []models.Metrics) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.SendCount++
}