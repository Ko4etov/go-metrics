package main

import (
	"flag"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	serverAddress := flag.String("a", "localhost:8080", "Server address")

	flag.Parse()

	r := chi.NewRouter()

	// Добавляем полезные middleware
	r.Use(middleware.Logger) // Логирование всех запросов

	// Объявляем маршруты
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handler.UpdateMetric)
	r.Get("/value/{metricType}/{metricName}", handler.GetMetric)
	r.Get("/", handler.GetMetrics)

	// Запускаем сервер
	err := http.ListenAndServe(*serverAddress, r)
	if err != nil {
		panic(err)
	}
}
