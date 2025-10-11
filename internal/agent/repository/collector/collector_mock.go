package collector

import (
	"github.com/Ko4etov/go-metrics/internal/models"
)

type CollectorMock struct {
	CollectCount int
}

func (m *CollectorMock) Collect() {
	m.CollectCount++
}

func (m *CollectorMock) Metrics() []models.Metrics {
	value := 1.0
	return []models.Metrics{
		{ID: "TestMetric", MType: models.Gauge, Value: &value},
	}
}

func (m *CollectorMock) PollCount() int64 {
	return int64(m.CollectCount)
}
