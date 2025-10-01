package handler

import (
	"fmt"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

func  GetMetricHandler(w http.ResponseWriter, r *http.Request) {
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

    // Получаем метрику из хранилища
    var value string
    var err error

    switch metricType {
    case "gauge":
        value, err = getGaugeMetric(metricName)
    case "counter":
        value, err = getCounterMetric(metricName)
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

// getGaugeMetric возвращает значение gauge метрики
func getGaugeMetric(name string) (string, error) {
	storage := storage.GetInstance()
    metric, exists := storage.GetMetric(name)
    if !exists || metric.MType != "gauge" {
        return "", fmt.Errorf("gauge metric not found")
    }
    
    if metric.Value == nil {
        return "", fmt.Errorf("invalid gauge value")
    }
    
    return fmt.Sprintf("%g", *metric.Value), nil
}

// getCounterMetric возвращает значение counter метрики
func getCounterMetric(name string) (string, error) {
    storage := storage.GetInstance()
    metric, exists := storage.GetMetric(name)
    if !exists || metric.MType != "counter" {
        return "", fmt.Errorf("counter metric not found")
    }
    
    if metric.Delta == nil {
        return "", fmt.Errorf("invalid counter value")
    }
    
    return fmt.Sprintf("%d", *metric.Delta), nil
}