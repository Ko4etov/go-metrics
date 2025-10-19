package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
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
		value, err = h.storage.GaugeMetric(metricName)
	case "counter":
		value, err = h.storage.CounterMetric(metricName)
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
