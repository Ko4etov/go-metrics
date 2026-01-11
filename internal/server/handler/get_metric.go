package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetMetric возвращает значение метрики в текстовом формате.
func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	if metricType == "" || metricName == "" {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

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

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, value)
}