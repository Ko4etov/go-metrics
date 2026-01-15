// Package collector реализует сбор метрик системы.
package collector

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/Ko4etov/go-metrics/internal/models"
)

// MetricsCollector реализует сбор и хранение метрик.
type MetricsCollector struct {
	mu          sync.RWMutex
	metrics     map[string]models.Metrics
	pollCounter int
	rand        *rand.Rand
}

// New создает новый сборщик метрик.
func New() *MetricsCollector {
	return &MetricsCollector{
		metrics:     make(map[string]models.Metrics),
		pollCounter: 0,
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Collect собирает метрики системы.
func (c *MetricsCollector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pollCounter++

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	runtimeMetrics := map[string]float64{
		"Alloc":         float64(stats.Alloc),
		"BuckHashSys":   float64(stats.BuckHashSys),
		"Frees":         float64(stats.Frees),
		"GCCPUFraction": stats.GCCPUFraction,
		"GCSys":         float64(stats.GCSys),
		"HeapAlloc":     float64(stats.HeapAlloc),
		"HeapIdle":      float64(stats.HeapIdle),
		"HeapInuse":     float64(stats.HeapInuse),
		"HeapObjects":   float64(stats.HeapObjects),
		"HeapReleased":  float64(stats.HeapReleased),
		"HeapSys":       float64(stats.HeapSys),
		"LastGC":        float64(stats.LastGC),
		"Lookups":       float64(stats.Lookups),
		"MCacheInuse":   float64(stats.MCacheInuse),
		"MCacheSys":     float64(stats.MCacheSys),
		"MSpanInuse":    float64(stats.MSpanInuse),
		"MSpanSys":      float64(stats.MSpanSys),
		"Mallocs":       float64(stats.Mallocs),
		"NextGC":        float64(stats.NextGC),
		"NumForcedGC":   float64(stats.NumForcedGC),
		"NumGC":         float64(stats.NumGC),
		"OtherSys":      float64(stats.OtherSys),
		"PauseTotalNs":  float64(stats.PauseTotalNs),
		"StackInuse":    float64(stats.StackInuse),
		"StackSys":      float64(stats.StackSys),
		"Sys":           float64(stats.Sys),
		"TotalAlloc":    float64(stats.TotalAlloc),
	}

	pollCount := c.pollCounter

	for name, value := range runtimeMetrics {
		valueCopy := value
		c.metrics[name] = models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &valueCopy,
		}
	}

	pollCountCopy := int64(pollCount)
	c.metrics["PollCount"] = models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &pollCountCopy,
	}

	randValue := rand.Float64() * 100
	c.metrics["RandomValue"] = models.Metrics{
		ID:    "RandomValue",
		MType: models.Gauge,
		Value: &randValue,
	}

	c.collectGopsutilMetrics()
}

// collectGopsutilMetrics собирает метрики системы через gopsutil.
func (c *MetricsCollector) collectGopsutilMetrics() {
	if memStats, err := mem.VirtualMemory(); err == nil {
		totalMemory := float64(memStats.Total)
		freeMemory := float64(memStats.Free)

		c.metrics["TotalMemory"] = models.Metrics{
			ID:    "TotalMemory",
			MType: models.Gauge,
			Value: &totalMemory,
		}

		c.metrics["FreeMemory"] = models.Metrics{
			ID:    "FreeMemory",
			MType: models.Gauge,
			Value: &freeMemory,
		}
	}

	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		cpuUtilization := cpuPercent[0]
		c.metrics["CPUutilization1"] = models.Metrics{
			ID:    "CPUutilization1",
			MType: models.Gauge,
			Value: &cpuUtilization,
		}
	}

	if cpuPercent, err := cpu.Percent(time.Second, true); err == nil {
		for i, percent := range cpuPercent {
			percentCopy := percent
			c.metrics[formatCPUutilization(i)] = models.Metrics{
				ID:    formatCPUutilization(i),
				MType: models.Gauge,
				Value: &percentCopy,
			}
		}
	}
}

// formatCPUutilization форматирует имя метрики загрузки CPU.
func formatCPUutilization(index int) string {
	return "CPUutilization" + strconv.Itoa(index+1)
}

// Metrics возвращает все собранные метрики.
func (c *MetricsCollector) Metrics() []models.Metrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(c.metrics))
	for _, metric := range c.metrics {
		metrics = append(metrics, metric)
	}

	return metrics
}

// PollCountReset сбрасывает счетчик опросов.
func (c *MetricsCollector) PollCountReset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pollCounter = 0
}

// PollCount возвращает текущее количество опросов.
func (c *MetricsCollector) PollCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.pollCounter
}