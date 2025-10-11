package metrics_sender

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/go-resty/resty/v2"
)

// Sender отправляет метрики на сервер
type MetricsSenderService struct {
    ServerAddress string
    Client        *http.Client
}

// NewSender создает новый отправитель
func New(serverAddress string) *MetricsSenderService {
    return &MetricsSenderService{
        ServerAddress: serverAddress,
        Client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// SendMetrics отправляет все метрики на сервер
func (s *MetricsSenderService) SendMetrics(metrics []models.Metrics) {
    for _, metric := range metrics {
        if err := s.SendMetric(metric); err != nil {
            log.Printf("Error sending metric %s: %v", metric.ID, err)
        } else {
            log.Printf("Metric sent successfully: %s = %v", metric.ID, metric.Value)
        }
    }
}

// sendMetric отправляет одну метрику на сервер
func (s *MetricsSenderService) SendMetric(metric models.Metrics) error {
    url := s.BuildURL(metric)
    
    // Создаем клиент с настройками
    client := resty.New().
        SetTimeout(5 * time.Second).
        SetRetryCount(2)
    
    resp, err := client.R().
        SetHeader("Content-Type", "text/plain").
        Post(url)
    
    if err != nil {
        return fmt.Errorf("send request failed: %w", err)
    }
    
    if resp.IsError() {
        return fmt.Errorf("server error: %s", resp.Status())
    }
    
    return nil
}

// buildURL строит URL для отправки метрики
func (s *MetricsSenderService) BuildURL(metric models.Metrics) string {
    var value string

    switch metric.MType {
    case "gauge":
		value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
    case "counter":
		value = strconv.FormatInt(*metric.Delta, 10)
    }

    return fmt.Sprintf("http://%s/update/%s/%s/%s", 
        s.ServerAddress, metric.MType, metric.ID, value)
}