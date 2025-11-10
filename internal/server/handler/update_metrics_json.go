package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/models"
)

func (h *Handler) UpdateMetricsBatch(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var metrics []models.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		http.Error(res, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		http.Error(res, "Empty metrics batch", http.StatusBadRequest)
		return
	}

	var validMetrics []models.Metrics
	for _, metric := range metrics {
		if err := h.validateMetric(&metric); err != nil {
			http.Error(res, "Invalid metric: "+err.Error(), http.StatusBadRequest)
			return
		}
		validMetrics = append(validMetrics, metric)
	}

	// Обновляем метрики в хранилище
	if err := h.storage.UpdateMetricsBatch(validMetrics); err != nil {
		http.Error(res, "Failed to update metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(map[string]string{"status": "ok", "updated": string(len(validMetrics))})
}