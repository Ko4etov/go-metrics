package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/server/config"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/Ko4etov/go-metrics/internal/server/router"
)

// Пример обновления метрики типа gauge через текстовый формат.
func Example_updateMetricText() {
	cfg := &config.ServerConfig{
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	routerConfig := &router.RouteConfig{
		Storage: store,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Post(
		fmt.Sprintf("%s/update/gauge/Alloc/123.45", server.URL),
		"text/plain",
		nil,
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output:
	// Status: 200
}

// Пример обновления метрики типа counter через текстовый формат.
func Example_updateCounterText() {
	cfg := &config.ServerConfig{
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	routerConfig := &router.RouteConfig{
		Storage: store,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Post(
		fmt.Sprintf("%s/update/counter/PollCount/1", server.URL),
		"text/plain",
		nil,
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output:
	// Status: 200
}

// Пример обновления метрики через JSON формат.
func Example_updateMetricJSON() {
	cfg := &config.ServerConfig{
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	routerConfig := &router.RouteConfig{
		Storage: store,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	value := 123.45
	metric := models.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &value,
	}

	jsonData, _ := json.Marshal(metric)

	resp, err := http.Post(
		fmt.Sprintf("%s/update/", server.URL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output:
	// Status: 200
}

// Пример получения метрики в текстовом формате.
func Example_getMetricText() {
	cfg := &config.ServerConfig{
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	// Добавляем тестовую метрику
	value := 123.45
	err := store.UpdateMetric(models.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &value,
	})
	if err != nil {
		fmt.Printf("Error updating metric: %v\n", err)
		return
	}

	routerConfig := &router.RouteConfig{
		Storage: store,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Get(fmt.Sprintf("%s/value/gauge/Alloc", server.URL))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var body bytes.Buffer
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Value: %s", strings.TrimSpace(body.String()))
	// Output:
	// Value: 123.45
}

// Пример получения метрики в JSON формате.
func Example_getMetricJSON() {
	cfg := &config.ServerConfig{
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	value := 123.45
	err := store.UpdateMetric(models.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &value,
	})
	if err != nil {
		fmt.Printf("Error updating metric: %v\n", err)
		return
	}

	routerConfig := &router.RouteConfig{
		Storage: store,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	reqMetric := models.Metrics{
		ID:    "Alloc",
		MType: "gauge",
	}
	jsonReq, _ := json.Marshal(reqMetric)

	resp, err := http.Post(
		fmt.Sprintf("%s/value/", server.URL),
		"application/json",
		bytes.NewBuffer(jsonReq),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result models.Metrics
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	fmt.Printf("ID: %s, Value: %.2f", result.ID, *result.Value)
	// Output:
	// ID: Alloc, Value: 123.45
}

// Пример обновления батча метрик.
func Example_updateBatchMetrics() {
	cfg := &config.ServerConfig{
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	routerConfig := &router.RouteConfig{
		Storage: store,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	// Подготовка батча метрик
	value := 123.45
	delta := int64(1)
	metrics := []models.Metrics{
		{
			ID:    "Alloc",
			MType: "gauge",
			Value: &value,
		},
		{
			ID:    "PollCount",
			MType: "counter",
			Delta: &delta,
		},
	}

	jsonData, _ := json.Marshal(metrics)

	resp, err := http.Post(
		fmt.Sprintf("%s/updates/", server.URL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output:
	// Status: 200
}

// Пример проверки соединения с базой данных.
// Этот пример показывает, что при отсутствии БД ping возвращает ошибку.
func Example_dBPing_noDB() {
	cfg := &config.ServerConfig{
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	routerConfig := &router.RouteConfig{
		Storage: store,
		Pgx:     nil, // Нет подключения к БД
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	// Проверка ping при отсутствии БД
	resp, err := http.Get(fmt.Sprintf("%s/ping", server.URL))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output:
	// Status: 500
}

// Пример получения всех метрик на HTML странице.
func Example_getAllMetrics() {
	cfg := &config.ServerConfig{
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	value1 := 100.5
	store.UpdateMetric(models.Metrics{
		ID:    "Metric1",
		MType: "gauge",
		Value: &value1,
	})

	value2 := int64(10)
	store.UpdateMetric(models.Metrics{
		ID:    "Metric2",
		MType: "counter",
		Delta: &value2,
	})

	mockPool := &pgxpool.Pool{}
	
	routerConfig := &router.RouteConfig{
		Storage: store,
		Pgx:     mockPool,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		fmt.Printf("Note: Требуется файл шаблона internal/server/templates/metrics.html")
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d", resp.StatusCode)
	// Output:
	// Status: 200
}

// Пример с накоплением значения counter метрики.
func Example_counterAccumulation() {
	cfg := &config.ServerConfig{
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	mockPool := &pgxpool.Pool{}
	
	routerConfig := &router.RouteConfig{
		Storage: store,
		Pgx:     mockPool,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	delta1 := int64(5)
	metric1 := models.Metrics{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &delta1,
	}

	jsonData1, _ := json.Marshal(metric1)
	resp1, err := http.Post(
		fmt.Sprintf("%s/update/", server.URL),
		"application/json",
		bytes.NewBuffer(jsonData1),
	)
	if err != nil {
		fmt.Printf("Error getting: %v\n", err)
		return
	}
	defer resp1.Body.Close()

	delta2 := int64(3)
	metric2 := models.Metrics{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &delta2,
	}

	jsonData2, _ := json.Marshal(metric2)
	resp2, err2 := http.Post(
		fmt.Sprintf("%s/update/", server.URL),
		"application/json",
		bytes.NewBuffer(jsonData2),
	)
	if err2 != nil {
		fmt.Printf("Error getting: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	resp3, err3 := http.Get(fmt.Sprintf("%s/value/counter/TestCounter", server.URL))
	if err3 != nil {
		fmt.Printf("Error getting: %v\n", err)
		return
	}
	defer resp3.Body.Close()

	var body bytes.Buffer
	_, err = body.ReadFrom(resp3.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Counter value: %s", strings.TrimSpace(body.String()))
	// Output:
	// Counter value: 8
}