// Package interfaces содержит определения интерфейсов для системы сбора метрик.
package interfaces

import "github.com/Ko4etov/go-metrics/internal/models"

// Collector определяет интерфейс сборщика метрик.
type Collector interface {
	Collect()
	Metrics() []models.Metrics
	PollCountReset()
	PollCount() int
}