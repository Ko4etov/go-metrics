package metricssender

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/go-resty/resty/v2"
)

// MetricsSenderService отправляет метрики на сервер с поддержкой хэширования
type MetricsSenderService struct {
	ServerAddress string
	HashKey       string
	Client        *resty.Client
	BatchSize     int
	MaxRetries    int
	RetryDelays   []time.Duration
}

// New создает новый отправитель
func New(serverAddress string, hashKey string) *MetricsSenderService {
	client := resty.New().
		SetTimeout(5 * time.Second).
		SetRetryCount(2)

	return &MetricsSenderService{
		ServerAddress: serverAddress,
		HashKey:       hashKey,
		Client:        client,
		BatchSize:     10,
		MaxRetries:    3,
		RetryDelays:   []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

// calculateHash вычисляет HMAC-SHA256 хеш для данных
func (s *MetricsSenderService) calculateHash(data []byte) string {
	if s.HashKey == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(s.HashKey))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// addHashHeaders добавляет заголовки хэширования к запросу
func (s *MetricsSenderService) addHashHeaders(req *resty.Request, data []byte) *resty.Request {
	if s.HashKey != "" && len(data) > 0 {
		hash := s.calculateHash(data)
		req.SetHeader("HashSHA256", hash)
		log.Printf("Added hash header for %d bytes of data", len(data))
		log.Printf("Data for hash: %s", string(data))
		log.Printf("Computed hash: %s", hash)
	}
	return req
}

// verifyResponseHash проверяет хеш ответа от сервера
func (s *MetricsSenderService) verifyResponseHash(resp *resty.Response) error {
	if s.HashKey == "" {
		return nil // Хеширование отключено
	}

	receivedHash := resp.Header().Get("HashSHA256")
	if receivedHash == "" {
		log.Printf("Warning: No hash header in response from server")
		return nil // Сервер может не отправлять хеш для некоторых ответов
	}

	// Вычисляем ожидаемый хеш от тела ответа
	expectedHash := s.calculateHash(resp.Body())
	if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
		return fmt.Errorf("response hash verification failed: received %s, expected %s",
			receivedHash, expectedHash)
	}

	log.Printf("Response hash verified successfully")
	return nil
}

// SendMetrics отправляет все метрики на сервер батчами с хэшированием
func (s *MetricsSenderService) SendMetrics(metrics []models.Metrics) {
	if len(metrics) == 0 {
		log.Printf("No metrics to send")
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
		if err := s.sendWithRetry(func() error {
			return s.SendMetricJSON(metric)
		}); err != nil {
			log.Printf("Error sending metric %s after %d retries: %v", metric.ID, s.MaxRetries, err)
		} else {
			log.Printf("Metric sent successfully: %s", metric.ID)
		}
	}
}

func (s *MetricsSenderService) sendBatchMetrics(metrics []models.Metrics) {
	batches := s.splitIntoBatches(metrics)

	for i, batch := range batches {
		log.Printf("Sending batch %d/%d with %d metrics", i+1, len(batches), len(batch))

		if err := s.sendWithRetry(func() error {
			return s.sendBatch(batch)
		}); err != nil {
			log.Printf("Error sending batch %d after %d retries: %v", i+1, s.MaxRetries, err)

			// При ошибке батча отправляем метрики по одной с retry
			s.sendSingleMetrics(batch)
		} else {
			log.Printf("Batch %d sent successfully", i+1)
		}
	}
}

func (s *MetricsSenderService) sendWithRetry(operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= s.MaxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Проверяем, является ли ошибка retriable
		if !s.isRetriableError(err) {
			return fmt.Errorf("non-retriable error: %w", err)
		}

		// Если это не последняя попытка, ждем перед повторной попыткой
		if attempt < s.MaxRetries {
			delay := s.RetryDelays[attempt]
			log.Printf("Attempt %d failed: %v. Retrying in %v", attempt+1, err, delay)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("failed after %d retries: %w", s.MaxRetries, lastErr)
}

func (s *MetricsSenderService) isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	// Ошибки хэширования не являются retriable
	if strings.Contains(strings.ToLower(err.Error()), "hash") {
		return false
	}

	// Проверяем типы ошибок
	if s.isNetworkError(err) {
		return true
	}

	// Проверяем по содержимому ошибки
	if s.isRetriableByContent(err) {
		return true
	}

	// Проверяем HTTP статусы
	if s.isRetriableHTTPStatus(err) {
		return true
	}

	return false
}

// isNetworkError проверяет сетевые ошибки
func (s *MetricsSenderService) isNetworkError(err error) bool {
	switch e := err.(type) {
	case *url.Error:
		return true
	case net.Error:
		return e.Timeout()
	}
	return false
}

// isRetriableByContent проверяет ошибки по их текстовому содержимому
func (s *MetricsSenderService) isRetriableByContent(err error) bool {
	errorStr := strings.ToLower(err.Error())

	retriablePatterns := []string{
		"timeout", "connection refused", "connection reset",
		"network", "temporary", "unavailable", "dial tcp",
		"no such host", "EOF", "broken pipe", "connection aborted",
		"i/o timeout", "network is unreachable", "reset by peer",
		"service unavailable", "bad gateway", "gateway timeout",
	}

	for _, pattern := range retriablePatterns {
		if strings.Contains(errorStr, pattern) {
			return true
		}
	}

	return false
}

// isRetriableHTTPStatus проверяет HTTP статусы ошибок
func (s *MetricsSenderService) isRetriableHTTPStatus(err error) bool {
	// Проверяем HTTP статусы через resty.Response если доступно
	if respErr, ok := err.(interface{ Response() *resty.Response }); ok {
		if resp := respErr.Response(); resp != nil {
			statusCode := resp.StatusCode()
			// 5xx ошибки и 429 (Too Many Requests) - retriable
			return statusCode >= 500 && statusCode <= 599 || statusCode == 429
		}
	}
	return false
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

// sendBatch отправляет один батч метрик на сервер с хэшированием
func (s *MetricsSenderService) sendBatch(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	url := fmt.Sprintf("http://%s/updates/", s.ServerAddress)

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics failed: %w", err)
	}

	compressedData, err := s.compressData(jsonData)
	if err != nil {
		return fmt.Errorf("compress data failed: %w", err)
	}

	req := s.Client.R().
		SetBody(compressedData).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip")

	req = s.addHashHeaders(req, jsonData)

	resp, err := req.Post(url)
	if err != nil {
		return fmt.Errorf("send batch request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server error: %s", resp.Status())
	}

	// Проверяем хеш ответа
	if err := s.verifyResponseHash(resp); err != nil {
		return fmt.Errorf("response hash verification failed: %w", err)
	}

	return nil
}

func (s *MetricsSenderService) marshalMetrics(metrics []models.Metrics) ([]byte, error) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("marshal metrics failed: %w", err)
	}
	return jsonData, nil
}

// compressData сжимает данные с помощью gzip
func (s *MetricsSenderService) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(data); err != nil {
		gz.Close()
		return nil, fmt.Errorf("gzip write failed: %w", err)
	}

	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("gzip close failed: %w", err)
	}

	return buf.Bytes(), nil
}

// SendMetric отправляет одну метрику текстовым форматом
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

// SendMetricJSON отправляет одну метрику JSON форматом с хэшированием
func (s *MetricsSenderService) SendMetricJSON(metric models.Metrics) error {
	url := fmt.Sprintf("http://%s/update/", s.ServerAddress)

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric failed: %w", err)
	}

	// Сжимаем данные
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		gz.Close()
		return fmt.Errorf("gzip write failed: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("gzip close failed: %w", err)
	}

	req := s.Client.R().
		SetBody(buf.Bytes()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Length", strconv.Itoa(buf.Len()))

	req = s.addHashHeaders(req, jsonData)

	resp, err := req.Post(url)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server error: %s", resp.Status())
	}

	// Проверяем хеш ответа
	if err := s.verifyResponseHash(resp); err != nil {
		return fmt.Errorf("response hash verification failed: %w", err)
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
