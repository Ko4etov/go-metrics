// Package router предоставляет маршрутизатор HTTP-запросов для сервера метрик.
package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ko4etov/go-metrics/internal/server/handler"
	"github.com/Ko4etov/go-metrics/internal/server/middlewares"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/Ko4etov/go-metrics/internal/server/service/audit"
)

// RouteConfig содержит конфигурацию для маршрутизатора.
type RouteConfig struct {
	Storage  *storage.MetricsStorage // хранилище метрик
	Pgx      *pgxpool.Pool           // пул подключений к базе данных
	HashKey  string                  // ключ для хеширования
	AuditSvc *audit.AuditService     // сервис аудита (опционально)
}

// New создает новый маршрутизатор с настройкой всех middleware и обработчиков.
func New(config *RouteConfig) *chi.Mux {
	metricHandler := handler.New(config.Storage, config.Pgx)
	hashConfig := &middlewares.HashConfig{
		SecretKey: config.HashKey,
	}

	r := chi.NewRouter()

	r.Use(middlewares.WithCompression)
	r.Use(middlewares.WithHashing(hashConfig))
	r.Use(middlewares.WithLogging)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.UpdateMetric)
	r.Post("/update/", metricHandler.UpdateMetricJSON)
	if config.AuditSvc != nil {
		auditHandler := metricHandler.UpdateMetricsBatchWithAudit(config.AuditSvc)
		r.Post("/updates/", auditHandler)
	} else {
		r.Post("/updates/", metricHandler.UpdateMetricsBatch)
	}
	r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetric)
	r.Post("/value/", metricHandler.GetMetricJSON)
	r.Get("/ping", metricHandler.DBPing)
	r.Get("/", metricHandler.GetMetrics)

	return r
}