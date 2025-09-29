package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ko4etov/go-metrics/internal/models"
)

func TestNewSender(t *testing.T) {
    sender := NewMetricsSenderService("localhost:8080")
    
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
    sender := NewMetricsSenderService("localhost:8080")
    
    value := 42.5
    metric := models.Metrics{
        ID:  "TestMetric",
        MType:  models.Gauge,
        Value: &value,
    }
    
    url := sender.BuildURL(metric)
    expected := "http://localhost:8080/update/gauge/TestMetric/42.5"
    
    if url != expected {
        t.Errorf("Expected URL %s, got %s", expected, url)
    }
}

func TestBuildURL_CounterMetric(t *testing.T) {
    sender := NewMetricsSenderService("localhost:8080")
    
    value := int64(100)

    metric := models.Metrics{
        ID:  "TestCounter",
        MType:  models.Counter,
        Delta: &value,
    }
    
    url := sender.BuildURL(metric)
    expected := "http://localhost:8080/update/counter/TestCounter/100"
    
    if url != expected {
        t.Errorf("Expected URL %s, got %s", expected, url)
    }
}

func TestBuildURL_EdgeCases(t *testing.T) {
    sender := NewMetricsSenderService("localhost:8080")

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
                ID:  "ZeroMetric",
                MType:  models.Gauge,
                Value: &zeroValue,
            },
            expected: "http://localhost:8080/update/gauge/ZeroMetric/0",
        },
        {
            name: "negative value",
            metric: models.Metrics{
                ID:  "NegativeMetric",
                MType:  models.Gauge,
                Value: &negativeValue,
            },
            expected: "http://localhost:8080/update/gauge/NegativeMetric/-10.5",
        },
        {
            name: "large counter",
            metric: models.Metrics{
                ID:  "LargeCounter",
                MType:  models.Counter,
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
    
    sender := NewMetricsSenderService(server.URL[7:]) // Убираем "http://"

    value := 42.5
    
    metric := models.Metrics{
        ID:  "TestMetric",
        MType:  models.Gauge,
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
    
    sender := NewMetricsSenderService(server.URL[7:])

    value := 42.5
    
    metric := models.Metrics{
        ID:  "TestMetric",
        MType:  models.Gauge,
        Value: &value,
    }
    
    err := sender.SendMetric(metric)
    if err == nil {
        t.Error("Expected error for server error response")
    }
}

func TestSendMetric_NetworkError(t *testing.T) {
    sender := NewMetricsSenderService("invalid-server:9999")

    value := 42.5
    
    metric := models.Metrics{
        ID:  "TestMetric",
        MType: models.Gauge,
        Value: &value,
    }
    
    err := sender.SendMetric(metric)
    if err == nil {
        t.Error("Expected error for network failure")
    }
}

func TestSendMetrics_MultipleMetrics(t *testing.T) {
    var requestCount int
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestCount++
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()
    
    sender := NewMetricsSenderService(server.URL[7:])

    value1 := 1.0
    value2 := int64(2)
    value3 := 3.0
    
    metrics := []models.Metrics{
        {ID: "Metric1", MType: models.Gauge, Value: &value1},
        {ID: "Metric2", MType: models.Counter, Delta: &value2},
        {ID: "Metric3", MType: models.Gauge, Value: &value3},
    }
    
    sender.SendMetrics(metrics)
    
    if requestCount != 3 {
        t.Errorf("Expected 3 requests, got %d", requestCount)
    }
}

func TestSendMetrics_PartialFailure(t *testing.T) {
    requestCount := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestCount++
        if requestCount == 2 {
            w.WriteHeader(http.StatusInternalServerError)
        } else {
            w.WriteHeader(http.StatusOK)
        }
    }))
    defer server.Close()
    
    sender := NewMetricsSenderService(server.URL[7:])

    value1 := 1.0
    value2 := int64(2)
    value3 := 3.0
    
    metrics := []models.Metrics{
        {ID: "Metric1", MType: models.Gauge, Value: &value1},
        {ID: "Metric2", MType: models.Counter, Delta: &value2},
        {ID: "Metric3", MType: models.Gauge, Value: &value3},
    }
    
    // Этот вызов не должен паниковать даже при ошибках
    sender.SendMetrics(metrics)
    
    if requestCount != 3 {
        t.Errorf("Expected 3 requests despite errors, got %d", requestCount)
    }
}