package router

import (
	"github.com/Ko4etov/go-metrics/internal/server/handler"
	"github.com/Ko4etov/go-metrics/internal/server/middlewares"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(metricsStorage *storage.MetricsStorage, pgx *pgxpool.Pool) *chi.Mux {
	metricHandler := handler.New(metricsStorage, pgx)

	r := chi.NewRouter()

	// Добавляем полезные middleware
	// r.Use(middleware.Logger) // Логирование всех запросов
	r.Use(middlewares.WithLoggingAndCompress)

	// Объявляем маршруты
	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.UpdateMetric)
	r.Post("/update/", metricHandler.UpdateMetricJSON)
	r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetric)
	r.Post("/value/", metricHandler.GetMetricJSON)
	r.Get("/ping", metricHandler.DbPing)
	r.Get("/", metricHandler.GetMetrics)

	return r;
}