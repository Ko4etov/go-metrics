package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Ko4etov/go-metrics/internal/server/config/db"
	"github.com/Ko4etov/go-metrics/internal/server/interfaces"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/go-chi/chi/v5"
)

func TestUpdateMetric(t *testing.T) {
	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics: false,
		StoreMetricsInterval: 100,
		FileStorageMetricsPath: "metrics.json",
	}
	storage := storage.New(storageConfig)
	poll := db.NewDbConnection("3456")
	metricHandler := New(storage, poll)
	storage.ResetAll()

	tests := []struct {
		name           string
		method         string
		metricType     string
		metricName     string
		metricValue    string
		contentType    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful gauge update",
			method:         http.MethodPost,
			metricType:     "gauge",
			metricName:     "temperature",
			metricValue:    "23.5",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "successful counter update",
			method:         http.MethodPost,
			metricType:     "counter",
			metricName:     "requests",
			metricValue:    "10",
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "invalid method GET",
			method:         http.MethodGet,
			metricType:     "gauge",
			metricName:     "test",
			metricValue:    "1.0",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "",
		},
		{
			name:           "invalid metric type",
			method:         http.MethodPost,
			metricType:     "invalid_type",
			metricName:     "test",
			metricValue:    "1.0",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Metric Type Not Allowed\n",
		},
		{
			name:           "empty metric name",
			method:         http.MethodPost,
			metricType:     "gauge",
			metricName:     "",
			metricValue:    "1.0",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric Not Found\n",
		},
		{
			name:           "invalid gauge value",
			method:         http.MethodPost,
			metricType:     "gauge",
			metricName:     "test",
			metricValue:    "not_a_float",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid gauge value\n",
		},
		{
			name:           "invalid counter value",
			method:         http.MethodPost,
			metricType:     "counter",
			metricName:     "test",
			metricValue:    "not_an_int",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid counter value\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/update/" + tt.metricType + "/" + tt.metricName + "/" + tt.metricValue
			req, err := http.NewRequest(tt.method, url, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.UpdateMetric)

			r.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if body := rr.Body.String(); body != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got %q want %q",
					body, tt.expectedBody)
			}

			if tt.expectedStatus == http.StatusOK {
				verifyMetricStored(t, storage, tt.metricType, tt.metricName, tt.metricValue)
			}
		})
	}
}

// verifyMetricStored checks if the metric was correctly stored
func verifyMetricStored(t *testing.T, storage interfaces.Storage, metricType, metricName, metricValue string) {
	t.Helper()

	metric, exists := storage.Metric(metricName)
	if !exists {
		t.Errorf("metric %s was not stored", metricName)
		return
	}

	if metric.MType != metricType {
		t.Errorf("metric type mismatch: got %s want %s", metric.MType, metricType)
	}

	switch metricType {
	case "gauge":
		if metric.Value == nil {
			t.Error("gauge value is nil")
			return
		}
		expectedValue, _ := strconv.ParseFloat(metricValue, 64)
		if *metric.Value != expectedValue {
			t.Errorf("gauge value mismatch: got %f want %f", *metric.Value, expectedValue)
		}
	case "counter":
		if metric.Delta == nil {
			t.Error("counter delta is nil")
			return
		}
		expectedValue, _ := strconv.ParseInt(metricValue, 10, 64)
		if *metric.Delta != expectedValue {
			t.Errorf("counter value mismatch: got %d want %d", *metric.Delta, expectedValue)
		}
	}
}
