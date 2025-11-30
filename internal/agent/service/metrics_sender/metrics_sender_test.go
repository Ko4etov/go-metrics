package metricssender

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Ko4etov/go-metrics/internal/models"
)

// Вспомогательные функции
func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestNewSender(t *testing.T) {
	sender := New("localhost:8080", "", 3)

	if sender == nil {
		t.Fatal("NewSender() returned nil")
	}

	if sender.ServerAddress != "localhost:8080" {
		t.Errorf("Expected serverAddress localhost:8080, got %s", sender.ServerAddress)
	}

	if sender.Client == nil {
		t.Error("HTTP client should be initialized")
	}
}

func TestBuildURL_GaugeMetric(t *testing.T) {
	sender := New("localhost:8080", "", 3)

	value := 42.5
	metric := models.Metrics{
		ID:    "TestMetric",
		MType: models.Gauge,
		Value: &value,
	}

	url := sender.BuildURL(metric)
	expected := "http://localhost:8080/update/gauge/TestMetric/42.5"

	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}
}

func TestBuildURL_CounterMetric(t *testing.T) {
	sender := New("localhost:8080", "", 3)

	value := int64(100)

	metric := models.Metrics{
		ID:    "TestCounter",
		MType: models.Counter,
		Delta: &value,
	}

	url := sender.BuildURL(metric)
	expected := "http://localhost:8080/update/counter/TestCounter/100"

	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}
}

func TestBuildURL_EdgeCases(t *testing.T) {
	sender := New("localhost:8080", "", 3)

	zeroValue := 0.0
	negativeValue := -10.5
	largeValue := int64(999999)

	tests := []struct {
		name     string
		metric   models.Metrics
		expected string
	}{
		{
			name: "zero values",
			metric: models.Metrics{
				ID:    "ZeroMetric",
				MType: models.Gauge,
				Value: &zeroValue,
			},
			expected: "http://localhost:8080/update/gauge/ZeroMetric/0",
		},
		{
			name: "negative value",
			metric: models.Metrics{
				ID:    "NegativeMetric",
				MType: models.Gauge,
				Value: &negativeValue,
			},
			expected: "http://localhost:8080/update/gauge/NegativeMetric/-10.5",
		},
		{
			name: "large counter",
			metric: models.Metrics{
				ID:    "LargeCounter",
				MType: models.Counter,
				Delta: &largeValue,
			},
			expected: "http://localhost:8080/update/counter/LargeCounter/999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := sender.BuildURL(tt.metric)
			if url != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, url)
			}
		})
	}
}

func TestSendMetric_Success(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Expected Content-Type text/plain, got %s", r.Header.Get("Content-Type"))
		}

		expectedPath := "/update/gauge/TestMetric/42.5"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := New(server.URL[7:], "", 3)

	value := 42.5

	metric := models.Metrics{
		ID:    "TestMetric",
		MType: models.Gauge,
		Value: &value,
	}

	err := sender.SendMetric(metric)
	if err != nil {
		t.Errorf("sendMetric failed: %v", err)
	}
}

func TestSendMetric_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := New(server.URL[7:], "", 3)

	value := 42.5

	metric := models.Metrics{
		ID:    "TestMetric",
		MType: models.Gauge,
		Value: &value,
	}

	err := sender.SendMetric(metric)
	if err == nil {
		t.Error("Expected error for server error response")
	}
}

func TestSendMetric_NetworkError(t *testing.T) {
	sender := New("invalid-server:9999", "", 3)

	value := 42.5

	metric := models.Metrics{
		ID:    "TestMetric",
		MType: models.Gauge,
		Value: &value,
	}

	err := sender.SendMetric(metric)
	if err == nil {
		t.Error("Expected error for network failure")
	}
}

func TestSendMetrics_MultipleMetrics(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := New(server.URL[7:], "", 3)

	value1 := 1.0
	value2 := int64(2)
	value3 := 3.0

	metrics := []models.Metrics{
		{ID: "Metric1", MType: models.Gauge, Value: &value1},
		{ID: "Metric2", MType: models.Counter, Delta: &value2},
		{ID: "Metric3", MType: models.Gauge, Value: &value3},
	}

	sender.SendMetrics(metrics)

	// Даем время worker'ам обработать
	time.Sleep(100 * time.Millisecond)

	// Останавливаем sender чтобы дождаться завершения worker'ов
	sender.Stop()

	// Теперь должно быть 1 batch запрос вместо 3 отдельных
	count := atomic.LoadInt32(&requestCount)
	if count != 1 {
		t.Errorf("Expected 1 batch request, got %d", count)
	}
}

func TestSendMetrics_PartialFailure(t *testing.T) {
	var requestCount int32
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		callCount++

		// Первый вызов падает, остальные успешны
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := New(server.URL[7:], "", 3)

	value1 := 1.0
	value2 := int64(2)
	value3 := 3.0

	metrics := []models.Metrics{
		{ID: "Metric1", MType: models.Gauge, Value: &value1},
		{ID: "Metric2", MType: models.Counter, Delta: &value2},
		{ID: "Metric3", MType: models.Gauge, Value: &value3},
	}

	sender.SendMetrics(metrics)

	// Даем время на retry логику
	time.Sleep(500 * time.Millisecond)

	// Останавливаем sender
	sender.Stop()

	// Проверяем количество запросов (должно быть несколько из-за retry)
	count := atomic.LoadInt32(&requestCount)
	if count < 2 {
		t.Errorf("Expected at least 2 requests due to retries, got %d", count)
	}
}

// Новый тест для проверки отправки по одной метрике
func TestSendMetricJSON_Success(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := New(server.URL[7:], "", 3)

	value := 42.5

	metric := models.Metrics{
		ID:    "TestMetric",
		MType: models.Gauge,
		Value: &value,
	}

	err := sender.SendMetricJSON(metric)
	if err != nil {
		t.Errorf("SendMetricJSON failed: %v", err)
	}

	if atomic.LoadInt32(&requestCount) != 1 {
		t.Errorf("Expected 1 request, got %d", atomic.LoadInt32(&requestCount))
	}
}