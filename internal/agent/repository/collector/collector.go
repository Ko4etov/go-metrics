package collector

import (
	"math/rand"
	"runtime"

	pollcounter "github.com/Ko4etov/go-metrics/internal/agent/service/poll_counter"
	"github.com/Ko4etov/go-metrics/internal/models"
)

type MetricsCollector struct {
	metrics   map[string]models.Metrics
	pollCounter *pollcounter.PollCounter
}

func New(pollCounter *pollcounter.PollCounter) *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]models.Metrics),
		pollCounter: pollCounter,
	}
}

func (c *MetricsCollector) Collect() {
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

	pollCount := c.pollCounter.Get()

	for name, value := range runtimeMetrics {
		c.metrics[name] = models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
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
}

func (c *MetricsCollector) Metrics() []models.Metrics {
	metrics := make([]models.Metrics, 0, len(c.metrics))
	for _, metric := range c.metrics {
		metrics = append(metrics, metric)
	}

	return metrics
}
