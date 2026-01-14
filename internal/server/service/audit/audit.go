// Package audit предоставляет систему аудита для отслеживания событий метрик.
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	retriableagent "github.com/Ko4etov/go-metrics/internal/service/retriable_agent"
	"github.com/go-resty/resty/v2"
)

// AuditEvent представляет событие аудита.
type AuditEvent struct {
	TS        int64    `json:"ts"`         // временная метка события
	Metrics   []string `json:"metrics"`    // список имен метрик
	IPAddress string   `json:"ip_address"` // IP-адрес клиента
}

// Auditor определяет интерфейс аудитора.
type Auditor interface {
	Audit(ctx context.Context, event AuditEvent) error
}

// Subscriber определяет интерфейс подписчика аудита.
type Subscriber interface {
	Auditor
	Name() string // имя подписчика
}

// AuditService управляет подписчиками и уведомляет их о событиях.
type AuditService struct {
	subscribers []Subscriber
	mu          sync.RWMutex
	enabled     bool
}

// NewAuditService создает новый сервис аудита.
func NewAuditService() *AuditService {
	return &AuditService{
		subscribers: make([]Subscriber, 0),
	}
}

// Subscribe добавляет подписчика к сервису аудита.
func (as *AuditService) Subscribe(subscriber Subscriber) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.subscribers = append(as.subscribers, subscriber)
	as.enabled = true
}

// Notify уведомляет всех подписчиков о событии аудита.
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

// FileAuditor реализует аудит в файл.
type FileAuditor struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

// NewFileAuditor создает новый файловый аудитор.
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

// Audit записывает событие аудита в файл.
func (fa *FileAuditor) Audit(ctx context.Context, event AuditEvent) error {

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	data = append(data, '\n')

	fa.mu.Lock()
	defer fa.mu.Unlock()

	if _, err := fa.file.Write(data); err != nil {
		return fmt.Errorf("failed to write to audit file: %w", err)
	}

	return nil
}

// Name возвращает имя файлового аудитора.
func (fa *FileAuditor) Name() string {
	return fmt.Sprintf("FileAuditor(%s)", fa.filePath)
}

// Close закрывает файловый аудитор.
func (fa *FileAuditor) Close() error {
	return fa.file.Close()
}

// HTTPAuditor реализует аудит по HTTP.
type HTTPAuditor struct {
	url           string
	Client        *resty.Client
	RetiebleAgent *retriableagent.RetriableAgent
}

// NewHTTPAuditor создает новый HTTP аудитор.
func NewHTTPAuditor(url string) *HTTPAuditor {
	client := resty.New().
		SetTimeout(5 * time.Second).
		SetRetryCount(2)

	retriableAgent := retriableagent.New(3, []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second})

	return &HTTPAuditor{
		url:           url,
		Client:        client,
		RetiebleAgent: retriableAgent,
	}
}

// Audit отправляет событие аудита по HTTP.
func (ha *HTTPAuditor) Audit(ctx context.Context, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	req := ha.Client.R().
		SetBody(data).
		SetHeader("Content-Type", "application/json")

	resp, err := req.Post(ha.url)
	if err != nil {
		return fmt.Errorf("failed to send audit event: %w", err)
	}
	
	if resp.IsError() {
		return fmt.Errorf("server error: %s", resp.Status())
	}

	return nil
}

// Name возвращает имя HTTP аудитора.
func (ha *HTTPAuditor) Name() string {
	return fmt.Sprintf("HTTPAuditor(%s)", ha.url)
}
