package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/storage"
)

func UpdateMetricHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.Header().Set("Allowed", http.MethodPost)
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if req.Header.Get("Content-Type") != "text/plain" {
		http.Error(res, "Content-Type Not Allowed", http.StatusBadRequest)
		return
	}

	metricType := req.PathValue("metric_type")
	metricName := req.PathValue("metric_name")
	metricValue := req.PathValue("metric_value")

	if metricType != "gauge" && metricType != "counter" {
		http.Error(res, "Metric Type Not Allowed", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(metricName) == "" {
		http.Error(res, "Metric Not Found", http.StatusNotFound)
		return
	}

	storage := storage.GetInstance()

	 var metric models.Metrics

	switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(res, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			metric = models.Metrics{
				ID:    metricName,
				MType: "gauge",
				Value: &value,
			}

		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(res, "Invalid counter value", http.StatusBadRequest)
				return
			}
			metric = models.Metrics{
				ID:    metricName,
				MType: "counter",
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

	fmt.Printf("metric_type = %s, metric_name = %s, metric_value = %s\n", metricType, metricName, metricValue)

	res.WriteHeader(http.StatusOK)
}