package repository

import "github.com/Ko4etov/go-metrics/internal/models"

type CollectorInterface interface {
	Collect()
	GetMetrics() []models.Metrics
	GetPollCount() int64
}