package metricssender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/go-resty/resty/v2"
)

// Sender отправляет метрики на сервер
type MetricsSenderService struct {
    ServerAddress string
    Client        *resty.Client
}

// NewSender создает новый отправитель
func New(serverAddress string) *MetricsSenderService {
    client := resty.New().
        SetTimeout(5 * time.Second).
        SetRetryCount(2)

    return &MetricsSenderService{
        ServerAddress: serverAddress,
        Client: client,
    }
}

// SendMetrics отправляет все метрики на сервер
func (s *MetricsSenderService) SendMetrics(metrics []models.Metrics) {
    for _, metric := range metrics {
        if err := s.SendMetricJSON(metric); err != nil {
            log.Printf("Error sending metric %s: %v", metric.ID, err)
        } else {
            log.Printf("Metric sent successfully: %s = %v", metric.ID, metric.Value)
        }
    }
}

// sendMetric отправляет одну метрику на сервер
func (s *MetricsSenderService) SendMetric(metric models.Metrics) error {
    url := s.BuildURL(metric)

    resp, err := s.Client.R().
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

// sendMetric отправляет одну метрику на сервер
func (s *MetricsSenderService) SendMetricJSON(metric models.Metrics) error {
    url := fmt.Sprintf("http://%s/update/", s.ServerAddress)

    jsonMetric, err := json.Marshal(metric)
    if err != nil {
        return err
    }
    var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer gz.Close()

	if _, err := gz.Write(jsonMetric); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
    
    resp, err := s.Client.R().
        SetBody(buf.Bytes()).
        SetHeader("Content-Type", "application/json").
        SetHeader("Content-Encoding", "gzip").
        SetHeader("Accept-Encoding", "gzip").
        SetHeader("Content-Length", strconv.Itoa(buf.Len())).
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