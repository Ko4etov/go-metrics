package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/models"
)

// UpdateMetricJSON обновляет метрику из JSON-запроса.
func (h *Handler) UpdateMetricJSON(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var metric models.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
		http.Error(res, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.validateMetric(&metric); err != nil {
		http.Error(res, "Invalid metric: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.storage.UpdateMetric(metric); err != nil {
		http.Error(res, "Failed to update metric: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(res).Encode(metric); err != nil {
		http.Error(res, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}