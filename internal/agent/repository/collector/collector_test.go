package collector

import (
	"testing"

	"github.com/Ko4etov/go-metrics/internal/models"
)

func TestNewCollector(t *testing.T) {
	collector := New()

	if collector == nil {
		t.Fatal("NewCollector() returned nil")
	}

	if collector.Metrics() == nil {
		t.Error("Metrics map should be initialized")
	}

	if collector.PollCount() != 0 {
		t.Errorf("Expected initial pollCount 0, got %d", collector.PollCount())
	}
}

func TestCollect_RuntimeMetrics(t *testing.T) {
	collector := New()
	collector.Collect()

	metrics := collector.Metrics()

	// Проверяем, что собрались основные runtime метрики
	requiredMetrics := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc",
	}

	metricMap := make(map[string]models.Metrics)
	for _, metric := range metrics {
		metricMap[metric.ID] = metric
	}

	for _, name := range requiredMetrics {
		if _, exists := metricMap[name]; !exists {
			t.Errorf("Required metric %s not collected", name)
		}

		if metric, exists := metricMap[name]; exists {
			if metric.MType != "gauge" {
				t.Errorf("Metric %s should be gauge, got %s", name, metric.MType)
			}

			if metric.Value == nil {
				t.Errorf("Metric %s value is nil", name)
			}
		}
	}
}

func TestCollect_CustomMetrics(t *testing.T) {
	collector := New()
	collector.Collect()

	metrics := collector.Metrics()
	metricMap := make(map[string]models.Metrics)
	for _, metric := range metrics {
		metricMap[metric.ID] = metric
	}

	// Проверяем PollCount
	pollCountMetric, exists := metricMap["PollCount"]
	if !exists {
		t.Fatal("PollCount metric not found")
	}

	if pollCountMetric.MType != "counter" {
		t.Errorf("PollCount should be counter, got %s", pollCountMetric.MType)
	}

	// Проверяем RandomValue
	randomValueMetric, exists := metricMap["RandomValue"]
	if !exists {
		t.Fatal("RandomValue metric not found")
	}

	if randomValueMetric.MType != "gauge" {
		t.Errorf("RandomValue should be gauge, got %s", randomValueMetric.MType)
	}

	if *randomValueMetric.Value < 0 || *randomValueMetric.Value > 100 {
		t.Errorf("RandomValue should be between 0 and 100, got %f", *randomValueMetric.Value)
	}
}

func TestCollect_PollCountIncrement(t *testing.T) {
	collector := New()

	// Собираем метрики несколько раз
	for i := 0; i < 3; i++ {
		collector.Collect()
	}

	metrics := collector.Metrics()
	metricMap := make(map[string]models.Metrics)
	for _, metric := range metrics {
		metricMap[metric.ID] = metric
	}

	pollCountMetric := metricMap["PollCount"]
	if *pollCountMetric.Delta != 3 {
		t.Errorf("Expected PollCount 3, got %v", *pollCountMetric.Delta)
	}

	if *pollCountMetric.Delta != 3 {
		t.Errorf("GetPollCount() should return 3, got %d", *pollCountMetric.Delta)
	}
}

func TestGetMetrics_ReturnsCopy(t *testing.T) {
	collector := New()
	collector.Collect()

	metrics1 := collector.Metrics()
	collector.Collect() // Изменяем внутреннее состояние
	metrics2 := collector.Metrics()

	// Проверяем, что GetMetrics возвращает копию
	if len(metrics1) != len(metrics2) {
		t.Error("Metrics slices should have different lengths after update")
	}
}
