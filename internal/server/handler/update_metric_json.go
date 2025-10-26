package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Ko4etov/go-metrics/internal/models"
)

func (h *Handler) UpdateMetricJson(res http.ResponseWriter, req *http.Request) {

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type Not Allowed Must Be application/json", http.StatusBadRequest)
	}

	var metric models.Metrics
	
	if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
		http.Error(res, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	if metric.MType != "gauge" && metric.MType != "counter" {
		http.Error(res, "Metric Type Not Allowed", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(metric.ID) == "" {
		http.Error(res, "Metric Not Found", http.StatusNotFound)
		return
	}

	if err := h.storage.UpdateMetric(metric); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("metric = %v\n", metric)


	res.WriteHeader(http.StatusOK)
}
