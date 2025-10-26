package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/models"
)

func (h *Handler) GetMetricJSON(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type Not Allowed Must Be application/json", http.StatusBadRequest)
	}

	var inputMetric models.Metrics
	
	if err := json.NewDecoder(req.Body).Decode(&inputMetric); err != nil {
		http.Error(res, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	
	// Валидация параметров
	if inputMetric.MType == "" || inputMetric.ID == "" {
		http.Error(res, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// Получаем метрику из хранилища
	var outputMetric *models.Metrics
	var err error

	switch inputMetric.MType {
		case "gauge":
			outputMetric, err = h.storage.GaugeMetricModel(inputMetric.ID)
		case "counter":
			outputMetric, err = h.storage.CounterMetricModel(inputMetric.ID)
		default:
			http.Error(res, "Invalid metric type", http.StatusBadRequest)
			return
	}

	if err != nil {
		http.Error(res, "Metric not found", http.StatusNotFound)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(res).Encode(outputMetric); err != nil {
		http.Error(res, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}
