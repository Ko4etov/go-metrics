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
	BatchSize     int
}

// New создает новый отправитель
func New(serverAddress string) *MetricsSenderService {
	client := resty.New().
		SetTimeout(5 * time.Second).
		SetRetryCount(2)

	return &MetricsSenderService{
		ServerAddress: serverAddress,
		Client:        client,
		BatchSize:     10,
	}
}

// SendMetrics отправляет все метрики на сервер батчами
func (s *MetricsSenderService) SendMetrics(metrics []models.Metrics) {
	if len(metrics) == 0 {
		log.Println("No metrics to send")
		return
	}

	if len(metrics) <= s.BatchSize {
		s.sendSingleMetrics(metrics)
		return
	}

	// Отправляем батчами
	s.sendBatchMetrics(metrics)
}

func (s *MetricsSenderService) sendSingleMetrics(metrics []models.Metrics) {
	for _, metric := range metrics {
		if err := s.SendMetricJSON(metric); err != nil {
			log.Printf("Error sending metric %s: %v", metric.ID, err)
		} else {
			log.Printf("Metric sent successfully: %s = %v", metric.ID, metric.Value)
		}
	}
}

func (s *MetricsSenderService) sendBatchMetrics(metrics []models.Metrics) {
	batches := s.splitIntoBatches(metrics)

	for i, batch := range batches {
		log.Printf("Sending batch %d/%d with %d metrics", i+1, len(batches), len(batch))
		
		if err := s.sendBatch(batch); err != nil {
			log.Printf("Error sending batch %d: %v", i+1, err)
			
			// При ошибке батча отправляем метрики по одной
			s.sendSingleMetrics(batch)
		} else {
			log.Printf("Batch %d sent successfully", i+1)
		}
	}
}

func (s *MetricsSenderService) splitIntoBatches(metrics []models.Metrics) [][]models.Metrics {
	var batches [][]models.Metrics

	for i := 0; i < len(metrics); i += s.BatchSize {
		end := i + s.BatchSize
		if end > len(metrics) {
			end = len(metrics)
		}
		batches = append(batches, metrics[i:end])
	}

	return batches
}

// sendBatch отправляет один батч метрик на сервер
func (s *MetricsSenderService) sendBatch(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	url := fmt.Sprintf("http://%s/updates/", s.ServerAddress)

	// Подготавливаем данные для отправки
	requestBody, err := s.prepareBatchRequestBody(metrics)
	if err != nil {
		return fmt.Errorf("prepare batch request failed: %w", err)
	}

	// Отправляем запрос
	resp, err := s.Client.R().
		SetBody(requestBody).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		Post(url)

	if err != nil {
		return fmt.Errorf("send batch request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server error: %s", resp.Status())
	}

	return nil
}

func (s *MetricsSenderService) prepareBatchRequestBody(metrics []models.Metrics) ([]byte, error) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("marshal metrics failed: %w", err)
	}

	// Сжимаем данные
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	
	if _, err := gz.Write(jsonData); err != nil {
		gz.Close()
		return nil, fmt.Errorf("gzip write failed: %w", err)
	}
	
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("gzip close failed: %w", err)
	}

	return buf.Bytes(), nil
}

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