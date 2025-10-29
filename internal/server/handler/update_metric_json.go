package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/models"
)

func (h *Handler) UpdateMetricJSON(res http.ResponseWriter, req *http.Request) {
    if req.Header.Get("Content-Type") != "application/json" {
        http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
        return
    }

    var metric models.Metrics
    if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
        http.Error(res, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if metric.MType != "gauge" && metric.MType != "counter" {
        http.Error(res, "Invalid metric type", http.StatusBadRequest)
        return
    }

    if err := h.storage.UpdateMetric(metric); err != nil {
        http.Error(res, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(res).Encode(metric)
}
