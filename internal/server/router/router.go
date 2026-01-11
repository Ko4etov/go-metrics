package router

import (
	"github.com/Ko4etov/go-metrics/internal/server/handler"
	"github.com/Ko4etov/go-metrics/internal/server/middlewares"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/Ko4etov/go-metrics/internal/server/service/audit"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouteConfig struct {
	Storage  *storage.MetricsStorage
	Pgx      *pgxpool.Pool
	HashKey  string
	AuditSvc *audit.AuditService
}

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
