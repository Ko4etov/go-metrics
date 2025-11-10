package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricsStorage struct {
	metrics    map[string]models.Metrics
	mu         *sync.Mutex
	config     *MetricsStorageConfig
	saveTicker *time.Ticker
	done       chan bool
}

type MetricsStorageConfig struct {
	RestoreMetrics         bool
	StoreMetricsInterval   int
	FileStorageMetricsPath string
	ConnectionPoll         *pgxpool.Pool
}

func New(config *MetricsStorageConfig) *MetricsStorage {
	storage := &MetricsStorage{
		metrics: make(map[string]models.Metrics),
		mu:      &sync.Mutex{},
		config:  config,
		done:    make(chan bool),
	}

	// Загрузка метрик при старте если нужно
	if config.RestoreMetrics {
		storage.LoadSavedMetrics()
	}

	return storage
}

func (ms *MetricsStorage) LoadSavedMetrics() {
	if ms.config.ConnectionPoll != nil {
		if err := ms.LoadFromDatabase(); err != nil {
			logger.Logger.Infof("Warning: failed to load metrics from database: %v\n", err)
		} else {
			logger.Logger.Infof("Successfully loaded metrics from database")
		}
	} else if ms.config.FileStorageMetricsPath != ""  {
		if err := ms.LoadFromFile(); err != nil {
			logger.Logger.Infof("Warning: failed to load metrics from file: %v\n", err)
		} else {
			logger.Logger.Infof("Successfully loaded metrics from %s\n", ms.config.FileStorageMetricsPath)
		}
	}
}

func (ms *MetricsStorage) LoadFromDatabase() error {
	ctx := context.Background()
	rows, err := ms.config.ConnectionPoll.Query(ctx, 
		"SELECT id, type, delta, value, hash FROM metrics")
	if err != nil {
		return fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	for rows.Next() {
		var metric models.Metrics
		err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value, &metric.Hash)
		if err != nil {
			return fmt.Errorf("failed to scan metric: %w", err)
		}
		ms.metrics[metric.ID] = metric
	}

	return rows.Err()
}

func (ms *MetricsStorage) StartPeriodicSave() {
	interval := time.Duration(ms.config.StoreMetricsInterval) * time.Second
	ms.saveTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ms.saveTicker.C:
				if err := ms.SaveToFile(); err != nil {
					fmt.Printf("Error saving metrics to file: %v\n", err)
				} else {
					fmt.Printf("Metrics automatically saved to %s\n", ms.config.FileStorageMetricsPath)
				}
			case <-ms.done:
				return
			}
		}
	}()
}

func (ms *MetricsStorage) SaveToFile() error {

	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Создаем директорию если нужно
	dir := filepath.Dir(ms.config.FileStorageMetricsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Конвертируем метрики в срез для JSON
	metricsSlice := make([]models.Metrics, 0, len(ms.metrics))
	for _, metric := range ms.metrics {
		metricsSlice = append(metricsSlice, metric)
	}

	// Сериализуем в JSON
	data, err := json.MarshalIndent(metricsSlice, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(ms.config.FileStorageMetricsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (ms *MetricsStorage) StopPeriodicSave() {
	if ms.saveTicker != nil {
		ms.saveTicker.Stop()
		close(ms.done)
	}
	
	// Финальное сохранение при остановке
	if ms.config.FileStorageMetricsPath != "" {
		if err := ms.SaveToFile(); err != nil {
			fmt.Printf("Error saving metrics on shutdown: %v\n", err)
		} else {
			fmt.Printf("Metrics saved to %s on shutdown\n", ms.config.FileStorageMetricsPath)
		}
	}
}

func (ms *MetricsStorage) LoadFromFile() error {
	if ms.config.FileStorageMetricsPath == "" {
		return errors.New("file storage path not specified")
	}

	// Проверяем существует ли файл
	if _, err := os.Stat(ms.config.FileStorageMetricsPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", ms.config.FileStorageMetricsPath)
	}

	// Читаем файл
	data, err := os.ReadFile(ms.config.FileStorageMetricsPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Декодируем JSON
	var metricsSlice []models.Metrics
	if err := json.Unmarshal(data, &metricsSlice); err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Загружаем метрики в хранилище
	for _, metric := range metricsSlice {
		ms.metrics[metric.ID] = metric
	}

	return nil
}

func (ms *MetricsStorage) Metrics() map[string]models.Metrics {
	return ms.metrics
}

func (ms *MetricsStorage) UpdateMetric(metric models.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return ErrInvalidValue
		}
		ms.metrics[metric.ID] = metric

	case models.Counter:
		if metric.Delta == nil {
			return ErrInvalidDelta
		}

		// Для counter добавляем значение к существующему
		if existing, exists := ms.metrics[metric.ID]; exists && existing.MType == models.Counter {
			newDelta := *existing.Delta + *metric.Delta
			metric.Delta = &newDelta
		}
		ms.metrics[metric.ID] = metric

	default:
		return ErrInvalidType
	}

	if ms.config.ConnectionPoll != nil {
		if err := ms.saveMetricToDatabase(metric); err != nil {
			return fmt.Errorf("failed to save metric to database: %w", err)
		}
	} else if ms.config.StoreMetricsInterval == 0 && ms.config.FileStorageMetricsPath != "" {
		// Или сохраняем в файл если нет БД
		return ms.SaveToFile()
	}

	return nil
}

func (ms *MetricsStorage) saveMetricToDatabase(metric models.Metrics) error {
	ctx := context.Background()
	
	_, err := ms.config.ConnectionPoll.Exec(ctx,
		`INSERT INTO metrics (id, type, delta, value, hash, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		 ON CONFLICT (id, type) 
		 DO UPDATE SET 
		   delta = EXCLUDED.delta,
		   value = EXCLUDED.value,
		   hash = EXCLUDED.hash,
		   updated_at = CURRENT_TIMESTAMP`,
		metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash)
	
	return err
}

func (ms *MetricsStorage) Metric(id string) (models.Metrics, bool) {
	metric, ok := ms.metrics[id]
	return metric, ok
}

func (ms *MetricsStorage) ResetAll() {
	ms.metrics = make(map[string]models.Metrics)
}

func (ms *MetricsStorage) GaugeMetric(name string) (string, error) {
	metric, exists := ms.Metric(name)
	if !exists || metric.MType != "gauge" {
		return "", fmt.Errorf("gauge metric not found")
	}

	if metric.Value == nil {
		return "", fmt.Errorf("invalid gauge value")
	}

	return fmt.Sprintf("%g", *metric.Value), nil
}

func (ms *MetricsStorage) CounterMetric(name string) (string, error) {
	metric, exists := ms.Metric(name)
	if !exists || metric.MType != "counter" {
		return "", fmt.Errorf("counter metric not found")
	}

	if metric.Delta == nil {
		return "", fmt.Errorf("invalid counter value")
	}

	return fmt.Sprintf("%d", *metric.Delta), nil
}

func (ms *MetricsStorage) GaugeMetricModel(name string) (*models.Metrics, error) {
	metric, exists := ms.Metric(name)
	if !exists || metric.MType != "gauge" {
		return &models.Metrics{}, fmt.Errorf("gauge metric not found")
	}

	if metric.Value == nil {
		return &models.Metrics{}, fmt.Errorf("invalid gauge value")
	}

	return &metric, nil
}

func (ms *MetricsStorage) CounterMetricModel(name string) (*models.Metrics, error) {
	metric, exists := ms.Metric(name)
	if !exists || metric.MType != "counter" {
		return &models.Metrics{}, fmt.Errorf("counter metric not found")
	}

	if metric.Delta == nil {
		return &models.Metrics{}, fmt.Errorf("invalid counter value")
	}

	return &metric, nil
}

func (ms *MetricsStorage) UpdateMetricsBatch(metrics []models.Metrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return ErrInvalidValue
			}
			ms.metrics[metric.ID] = metric

		case models.Counter:
			if metric.Delta == nil {
				return ErrInvalidDelta
			}

			if existing, exists := ms.metrics[metric.ID]; exists && existing.MType == models.Counter {
				newDelta := *existing.Delta + *metric.Delta
				metric.Delta = &newDelta
			}
			ms.metrics[metric.ID] = metric

		default:
			return ErrInvalidType
		}
	}

	if ms.config.ConnectionPoll != nil {
		if err := ms.saveMetricsBatchToDatabase(metrics); err != nil {
			return fmt.Errorf("failed to save metrics batch to database: %w", err)
		}
	} else if ms.config.StoreMetricsInterval == 0 && ms.config.FileStorageMetricsPath != "" {
		return ms.SaveToFile()
	}

	return nil
}

func (ms *MetricsStorage) saveMetricsBatchToDatabase(metrics []models.Metrics) error {
	ctx := context.Background()
	
	tx, err := ms.config.ConnectionPoll.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, metric := range metrics {
		_, err := tx.Exec(ctx,
			`INSERT INTO metrics (id, type, delta, value, hash, updated_at) 
			 VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
			 ON CONFLICT (id, type) 
			 DO UPDATE SET 
			   delta = EXCLUDED.delta,
			   value = EXCLUDED.value,
			   hash = EXCLUDED.hash,
			   updated_at = CURRENT_TIMESTAMP`,
			metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash)
		
		if err != nil {
			return fmt.Errorf("failed to save metric %s: %w", metric.ID, err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

var (
	ErrInvalidType  = errors.New("invalid metric type")
	ErrInvalidValue = errors.New("invalid value for gauge metric")
	ErrInvalidDelta = errors.New("invalid delta for counter metric")
)
