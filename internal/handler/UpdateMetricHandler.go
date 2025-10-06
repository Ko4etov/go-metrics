package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/repository/storage"
	"github.com/go-chi/chi/v5"
)

func UpdateMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.Header().Set("Allowed", http.MethodPost)
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	metricType := chi.URLParam(req, "metricType")
	metricName := chi.URLParam(req, "metricName")
	metricValue := chi.URLParam(req, "metricValue")

	if metricType != "gauge" && metricType != "counter" {
		http.Error(res, "Metric Type Not Allowed", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(metricName) == "" {
		http.Error(res, "Metric Not Found", http.StatusNotFound)
		return
	}

	storage := storage.New()

	var metric models.Metrics

	switch metricType {
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(res, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		metric = models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Value: &value,
		}

	case models.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(res, "Invalid counter value", http.StatusBadRequest)
			return
		}
		metric = models.Metrics{
			ID:    metricName,
			MType: models.Counter,
			Delta: &value,
		}

	default:
		http.Error(res, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if err := storage.UpdateMetric(metric); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("metric_type = %T, metric_name = %T, metric_value = %T\n", metricType, metricName, metricValue)

	res.WriteHeader(http.StatusOK)
}
