package router

import (
	"github.com/Ko4etov/go-metrics/internal/server/handler"
	"github.com/Ko4etov/go-metrics/internal/server/middlewares"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouteConfig struct {
	Storage *storage.MetricsStorage
	Pgx *pgxpool.Pool
	HashKey string
}

func New(config *RouteConfig) *chi.Mux {
	metricHandler := handler.New(config.Storage, config.Pgx)
	hashConfig := &middlewares.HashConfig{
		SecretKey: config.HashKey,
	}

	r := chi.NewRouter()

	// Добавляем полезные middleware
	r.Use(middlewares.WithHashing(hashConfig))
	r.Use(middlewares.WithCompression)
	r.Use(middlewares.WithLogging)

	// Объявляем маршруты
	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.UpdateMetric)
	r.Post("/update/", metricHandler.UpdateMetricJSON)
	r.Post("/updates/", metricHandler.UpdateMetricsBatch)
	r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetric)
	r.Post("/value/", metricHandler.GetMetricJSON)
	r.Get("/ping", metricHandler.DBPing)
	r.Get("/", metricHandler.GetMetrics)

	return r;
}