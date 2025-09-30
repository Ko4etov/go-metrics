package service

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Ko4etov/go-metrics/internal/models"
)

// Sender отправляет метрики на сервер
type MetricsSenderService struct {
    ServerAddress string
    Client        *http.Client
}

// NewSender создает новый отправитель
func NewMetricsSenderService(serverAddress string) *MetricsSenderService {
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

    log.Printf("%s", url)
    
    req, err := http.NewRequest("POST", url, nil)
    if err != nil {
        return fmt.Errorf("create request failed: %w", err)
    }

    req.Header.Set("Content-Type", "text/plain")

    resp, err := s.Client.Do(req)
    if err != nil {
        return fmt.Errorf("send request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("server returned status: %d", resp.StatusCode)
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

// SendMetricBatch отправляет метрики батчем (альтернативный метод)
func (s *MetricsSenderService) SendMetricBatch(metrics []models.Metrics) error {
    // Этот метод можно использовать для отправки метрик в одном запросе
    // если сервер поддерживает batch updates
    for _, metric := range metrics {
        if err := s.SendMetric(metric); err != nil {
            return err
        }
    }
    return nil
}