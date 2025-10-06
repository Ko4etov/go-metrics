package handler

import (
	"fmt"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/repository/storage"
	"github.com/go-chi/chi/v5"
)

func GetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем параметры из URL
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	// Валидация параметров
	if metricType == "" || metricName == "" {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	storage := storage.New()

	// Получаем метрику из хранилища
	var value string
	var err error

	switch metricType {
	case "gauge":
		value, err = storage.GaugeMetric(metricName)
	case "counter":
		value, err = storage.CounterMetric(metricName)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	// Возвращаем значение в текстовом виде
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, value)
}
