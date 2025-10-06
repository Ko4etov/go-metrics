package service

import "github.com/Ko4etov/go-metrics/internal/models"

type MetricsSenderInterface interface {
	SendMetrics(metrics []models.Metrics)
}
