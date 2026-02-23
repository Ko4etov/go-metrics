// Package metricssender предоставляет функциональность отправки метрик на сервер.
package metricssender

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/service/crypto"
	retriableagent "github.com/Ko4etov/go-metrics/internal/service/retriable_agent"
)

// MetricsSenderService отправляет метрики на сервер с поддержкой хэширования.
// generate:reset
type MetricsSenderService struct {
	ServerAddress string
	HashKey       string
	CryptoKey     *rsa.PublicKey
	Client        *resty.Client
	BatchSize     int
	RateLimit     int
	RetiebleAgent *retriableagent.RetriableAgent
	jobs          chan []models.Metrics
	wg            sync.WaitGroup
}

// New создает новый отправитель метрик.
func New(serverAddress string, hashKey string, rateLimit int, cryptoKey string) *MetricsSenderService {
	client := resty.New().
		SetTimeout(5 * time.Second).
		SetRetryCount(2)

	retriableAgent := retriableagent.New(3, []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second})

	sender := &MetricsSenderService{
		ServerAddress: serverAddress,
		HashKey:       hashKey,
		CryptoKey:     nil,
		Client:        client,
		BatchSize:     10,
		RetiebleAgent: retriableAgent,
		RateLimit:     rateLimit,
		jobs:          make(chan []models.Metrics, rateLimit),
	}

	if cryptoKey != "" {
		publicKey, err := crypto.LoadPublicKey(cryptoKey)
		if err != nil {
			fmt.Printf("Warning: failed to load public key from %s: %v\n", cryptoKey, err)
		} else {
			sender.CryptoKey = publicKey
		}
	}

	sender.startWorkers()

	return sender
}

func (s *MetricsSenderService) startWorkers() {
	for i := 0; i < s.RateLimit; i++ {
		s.wg.Add(1)
		go s.worker()
	}
}

func (s *MetricsSenderService) worker() {
	defer s.wg.Done()

	for metrics := range s.jobs {
		s.RetiebleAgent.Send(func() error {
			return s.sendBatch(metrics)
		})
	}
}

func (s *MetricsSenderService) calculateHash(data []byte) string {
	if s.HashKey == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(s.HashKey))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *MetricsSenderService) addHashHeaders(req *resty.Request, data []byte) *resty.Request {
	if s.HashKey != "" && len(data) > 0 {
		hash := s.calculateHash(data)
		req.SetHeader("HashSHA256", hash)
	}
	return req
}

func (s *MetricsSenderService) verifyResponseHash(resp *resty.Response) error {
	if s.HashKey == "" {
		return nil // Хеширование отключено
	}

	receivedHash := resp.Header().Get("HashSHA256")
	if receivedHash == "" {
		return nil // Сервер может не отправлять хеш для некоторых ответов
	}

	expectedHash := s.calculateHash(resp.Body())
	if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
		return fmt.Errorf("response hash verification failed: received %s, expected %s",
			receivedHash, expectedHash)
	}

	return nil
}

// SendMetrics отправляет метрики на сервер.
func (s *MetricsSenderService) SendMetrics(metrics []models.Metrics) {
	if len(metrics) == 0 {
		return
	}

	batches := s.splitIntoBatches(metrics)

	for _, batch := range batches {
		select {
		case s.jobs <- batch:
		default:
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

// sendBatch отправляет один батч метрик на сервер с хэшированием.
func (s *MetricsSenderService) sendBatch(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	url := fmt.Sprintf("http://%s/updates/", s.ServerAddress)

	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("Warning: failed to get local IP: %v", err)
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics failed: %w", err)
	}

	compressedData, err := s.compressData(jsonData)
	if err != nil {
		return fmt.Errorf("compress data failed: %w", err)
	}

	var finalData []byte
    var contentType string
    var contentEncoding string

    if s.CryptoKey != nil {
        encryptedData, err := crypto.Encrypt(compressedData, s.CryptoKey)
        if err != nil {
            return fmt.Errorf("encrypt data failed: %w", err)
        }
        finalData = encryptedData
        contentType = "application/octet-stream"
        contentEncoding = ""
    } else {
        finalData = compressedData
        contentType = "application/json"
        contentEncoding = "gzip"
    }

	req := s.Client.R().
        SetBody(finalData).
        SetHeader("Content-Type", contentType)

	if localIP != "" {
		req.SetHeader("X-Real-IP", localIP)
	}

    if contentEncoding != "" {
        req.SetHeader("Content-Encoding", contentEncoding)
    }
    req.SetHeader("Accept-Encoding", "gzip")

	req = s.addHashHeaders(req, jsonData)

	resp, err := req.Post(url)
	if err != nil {
		return fmt.Errorf("send batch request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server error: %s", resp.Status())
	}

	if err := s.verifyResponseHash(resp); err != nil {
		return fmt.Errorf("response hash verification failed: %w", err)
	}

	return nil
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		// Проверяем, что это IP адрес и не loopback
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	// Если не нашли, пробуем получить через hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	hostsAddrs, err := net.LookupHost(hostname)
	if err != nil {
		return "", err
	}

	if len(hostsAddrs) > 0 {
		return hostsAddrs[0], nil
	}

	return "", fmt.Errorf("could not determine local IP")
}

// compressData сжимает данные с помощью gzip.
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

// SendMetric отправляет одну метрику текстовым форматом.
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

// SendMetricJSON отправляет одну метрику JSON форматом с хэшированием.
func (s *MetricsSenderService) SendMetricJSON(metric models.Metrics) error {
	url := fmt.Sprintf("http://%s/update/", s.ServerAddress)

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("marshal metric failed: %w", err)
	}

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

	if err := s.verifyResponseHash(resp); err != nil {
		return fmt.Errorf("response hash verification failed: %w", err)
	}

	return nil
}

// BuildURL строит URL для отправки одной метрики текстовым форматом.
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

// Stop останавливает отправщик метрик и дожидается завершения всех воркеров.
func (s *MetricsSenderService) Stop() {
	close(s.jobs)
	s.wg.Wait()
}
