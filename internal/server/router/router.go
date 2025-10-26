package router

import (
	"github.com/Ko4etov/go-metrics/internal/server/handler"
	"github.com/Ko4etov/go-metrics/internal/server/middlewares"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New() *chi.Mux {
	metricsStorage := storage.New()
	
	metricHandler := handler.New(metricsStorage)

	r := chi.NewRouter()

	// Добавляем полезные middleware
	r.Use(middleware.Logger) // Логирование всех запросов
	r.Use(middlewares.WithLogging)

	// Объявляем маршруты
	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.UpdateMetric)
	r.Post("/update", metricHandler.UpdateMetricJSON)
	r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetric)
	r.Get("/", metricHandler.GetMetrics)

	return r;
}