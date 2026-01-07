package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEvent - событие аудита
type AuditEvent struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// Auditor - интерфейс аудитора
type Auditor interface {
	Audit(ctx context.Context, event AuditEvent) error
}

// Subscriber - интерфейс подписчика
type Subscriber interface {
	Auditor
	Name() string
}

// AuditService - сервис аудита (Subject в паттерне Observer)
type AuditService struct {
	subscribers []Subscriber
	mu          sync.RWMutex
	enabled     bool
}

// NewAuditService создает новый сервис аудита
func NewAuditService() *AuditService {
	return &AuditService{
		subscribers: make([]Subscriber, 0),
	}
}

// Subscribe добавляет подписчика
func (as *AuditService) Subscribe(subscriber Subscriber) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.subscribers = append(as.subscribers, subscriber)
	as.enabled = true
}

// Notify отправляет событие всем подписчикам
func (as *AuditService) Notify(ctx context.Context, event AuditEvent) error {
	if !as.enabled {
		return nil
	}

	as.mu.RLock()
	subscribers := make([]Subscriber, len(as.subscribers))
	copy(subscribers, as.subscribers)
	as.mu.RUnlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(subscribers))

	for _, sub := range subscribers {
		wg.Add(1)
		go func(s Subscriber) {
			defer wg.Done()
			if err := s.Audit(ctx, event); err != nil {
				errCh <- fmt.Errorf("subscriber %s: %w", s.Name(), err)
			}
		}(sub)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("audit failed: %v", errs)
	}
	return nil
}

// FileAuditor - аудитор для записи в файл
type FileAuditor struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

func NewFileAuditor(filePath string) (*FileAuditor, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit file: %w", err)
	}

	return &FileAuditor{
		filePath: filePath,
		file:     file,
	}, nil
}

func (fa *FileAuditor) Audit(ctx context.Context, event AuditEvent) error {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	data = append(data, '\n')

	if _, err := fa.file.Write(data); err != nil {
		return fmt.Errorf("failed to write to audit file: %w", err)
	}

	return nil
}

func (fa *FileAuditor) Name() string {
	return fmt.Sprintf("FileAuditor(%s)", fa.filePath)
}

func (fa *FileAuditor) Close() error {
	return fa.file.Close()
}

// HTTPAuditor - аудитор для отправки по HTTP
type HTTPAuditor struct {
	url    string
	client *http.Client
}

func NewHTTPAuditor(url string) *HTTPAuditor {
	return &HTTPAuditor{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (ha *HTTPAuditor) Audit(ctx context.Context, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ha.url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ha.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send audit event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("audit server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (ha *HTTPAuditor) Name() string {
	return fmt.Sprintf("HTTPAuditor(%s)", ha.url)
}