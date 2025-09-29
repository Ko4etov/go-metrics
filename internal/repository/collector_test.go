package repository

import (
	"testing"

	"github.com/Ko4etov/go-metrics/internal/models"
)

func TestNewCollector(t *testing.T) {
    collector := NewCollector()
    
    if collector == nil {
        t.Fatal("NewCollector() returned nil")
    }
    
    if collector.metrics == nil {
        t.Error("Metrics map should be initialized")
    }
    
    if collector.pollCount != 0 {
        t.Errorf("Expected initial pollCount 0, got %d", collector.pollCount)
    }
}

func TestCollect_RuntimeMetrics(t *testing.T) {
    collector := NewCollector()
    collector.Collect()
    
    metrics := collector.GetMetrics()
    
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
    collector := NewCollector()
    collector.Collect()
    
    metrics := collector.GetMetrics()
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
    
    if *pollCountMetric.Delta != 1 {
        t.Errorf("Expected PollCount 1, got %v", pollCountMetric.Value)
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
    collector := NewCollector()
    
    // Собираем метрики несколько раз
    for i := 0; i < 3; i++ {
        collector.Collect()
    }
    
    metrics := collector.GetMetrics()
    metricMap := make(map[string]models.Metrics)
    for _, metric := range metrics {
        metricMap[metric.ID] = metric
    }
    
    pollCountMetric := metricMap["PollCount"]
    if *pollCountMetric.Value != 3 {
        t.Errorf("Expected PollCount 3, got %v", pollCountMetric.Value)
    }
    
    if collector.GetPollCount() != 3 {
        t.Errorf("GetPollCount() should return 3, got %d", collector.GetPollCount())
    }
}

func TestCollect_ConcurrentSafety(t *testing.T) {
    collector := NewCollector()
    
    // Запускаем несколько горутин для конкурентного доступа
    done := make(chan bool)
    
    for i := 0; i < 5; i++ {
        go func() {
            for j := 0; j < 100; j++ {
                collector.Collect()
                collector.GetMetrics()
                collector.GetPollCount()
            }
            done <- true
        }()
    }
    
    // Ждем завершения всех горутин
    for i := 0; i < 5; i++ {
        <-done
    }
    
    // Проверяем, что нет паники и данные консистентны
    pollCount := collector.GetPollCount()
    if pollCount != 500 {
        t.Errorf("Expected pollCount 500, got %d", pollCount)
    }
    
    metrics := collector.GetMetrics()
    if len(metrics) == 0 {
        t.Error("No metrics collected")
    }
}

func TestGetMetrics_ReturnsCopy(t *testing.T) {
    collector := NewCollector()
    collector.Collect()
    
    metrics1 := collector.GetMetrics()
    collector.Collect() // Изменяем внутреннее состояние
    metrics2 := collector.GetMetrics()
    
    // Проверяем, что GetMetrics возвращает копию
    if len(metrics1) != len(metrics2) {
        t.Error("Metrics slices should have different lengths after update")
    }
}