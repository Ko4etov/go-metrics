package interfaces

import "github.com/Ko4etov/go-metrics/internal/models"

type Collector interface {
	Collect()
	Metrics() []models.Metrics
	PollCountReset()
	PollCount() int
}
