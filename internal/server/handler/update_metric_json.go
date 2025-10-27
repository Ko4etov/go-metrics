package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

func (h *Handler) UpdateMetricJSON(res http.ResponseWriter, req *http.Request) {

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type Not Allowed Must Be application/json", http.StatusBadRequest)
	}

	body, readErr := io.ReadAll(req.Body)

	logger.Logger.Infoln(
		"body", body,
	)

	if readErr != nil {
        http.Error(res, "Error reading request body", http.StatusInternalServerError)
        return
    }

	if len(body) == 0 {
        http.Error(res, "Empty request body", http.StatusBadRequest)
        return
    }

    if !json.Valid(body) {
        http.Error(res, "Invalid JSON format", http.StatusBadRequest)
        return
    }

	var metric models.Metrics
	
	if err := json.Unmarshal(body, &metric); err != nil {
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


	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(res).Encode(metric); err != nil {
		http.Error(res, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}
