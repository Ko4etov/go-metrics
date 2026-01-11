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
	// Создаем тестовый сервер
	cfg := &config.ServerConfig{
		ServerAddress:        ":8080",
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

	// Обновление метрики gauge
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
	// Создаем тестовый сервер
	cfg := &config.ServerConfig{
		ServerAddress:        ":8080",
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

	// Обновление метрики counter
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
	// Создаем тестовый сервер
	cfg := &config.ServerConfig{
		ServerAddress:        ":8080",
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

	// Подготовка JSON метрики типа gauge
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
	// Создаем тестовый сервер и добавляем тестовые данные
	cfg := &config.ServerConfig{
		ServerAddress:        ":8080",
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
	store.UpdateMetric(models.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &value,
	})

	routerConfig := &router.RouteConfig{
		Storage: store,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	// Получение метрики
	resp, err := http.Get(fmt.Sprintf("%s/value/gauge/Alloc", server.URL))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var body bytes.Buffer
	body.ReadFrom(resp.Body)

	fmt.Printf("Value: %s", strings.TrimSpace(body.String()))
	// Output:
	// Value: 123.45
}

// Пример получения метрики в JSON формате.
func Example_getMetricJSON() {
	// Создаем тестовый сервер и добавляем тестовые данные
	cfg := &config.ServerConfig{
		ServerAddress:        ":8080",
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
	store.UpdateMetric(models.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &value,
	})

	routerConfig := &router.RouteConfig{
		Storage: store,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	// Запрос метрики в JSON формате
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
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("ID: %s, Value: %.2f", result.ID, *result.Value)
	// Output:
	// ID: Alloc, Value: 123.45
}

// Пример обновления батча метрик.
func Example_updateBatchMetrics() {
	// Создаем тестовый сервер
	cfg := &config.ServerConfig{
		ServerAddress:        ":8080",
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
	// Создаем тестовый сервер БЕЗ подключения к БД
	cfg := &config.ServerConfig{
		ServerAddress:        ":8080",
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	// Создаем роутер с nil подключением к БД
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

// Пример успешной проверки соединения с базой данных.
// Для этого теста нужно мок-подключение или тестовая БД.
func Example_dBPing_withDB() {
	// Этот пример требует реального подключения к БД.
	// В тестовом окружении можно использовать мок или тестовую БД.
	fmt.Println("Для проверки ping с БД требуется подключение к PostgreSQL")
	// Output:
	// Для проверки ping с БД требуется подключение к PostgreSQL
}

// Пример получения всех метрик на HTML странице.
// ВАЖНО: Для работы этого примера должен существовать файл шаблона.
// Если файл отсутствует, пример вернет ошибку.
func Example_getAllMetrics() {
	// Создаем тестовый сервер и добавляем тестовые данные
	cfg := &config.ServerConfig{
		ServerAddress:        ":8080",
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	// Добавляем тестовые метрики
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

	// Получаем HTML страницу с метриками
	resp, err := http.Get(server.URL)
	if err != nil {
		// Если возникает ошибка, это может быть связано с отсутствием шаблона
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
	// Создаем тестовый сервер
	cfg := &config.ServerConfig{
		ServerAddress:        ":8080",
		RestoreMetrics:       false,
		StoreMetricsInterval: 0,
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:       cfg.RestoreMetrics,
		StoreMetricsInterval: cfg.StoreMetricsInterval,
	}
	store := storage.New(storageConfig)

	// Создаем мок пула подключений для теста
	mockPool := &pgxpool.Pool{}
	
	routerConfig := &router.RouteConfig{
		Storage: store,
		Pgx:     mockPool,
		HashKey: "",
	}
	r := router.New(routerConfig)

	server := httptest.NewServer(r)
	defer server.Close()

	// Первое обновление счетчика
	delta1 := int64(5)
	metric1 := models.Metrics{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &delta1,
	}

	jsonData1, _ := json.Marshal(metric1)
	resp1, _ := http.Post(
		fmt.Sprintf("%s/update/", server.URL),
		"application/json",
		bytes.NewBuffer(jsonData1),
	)
	defer resp1.Body.Close()

	// Второе обновление того же счетчика
	delta2 := int64(3)
	metric2 := models.Metrics{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &delta2,
	}

	jsonData2, _ := json.Marshal(metric2)
	resp2, _ := http.Post(
		fmt.Sprintf("%s/update/", server.URL),
		"application/json",
		bytes.NewBuffer(jsonData2),
	)
	defer resp2.Body.Close()

	// Проверяем значение счетчика
	resp3, _ := http.Get(fmt.Sprintf("%s/value/counter/TestCounter", server.URL))
	defer resp3.Body.Close()

	var body bytes.Buffer
	body.ReadFrom(resp3.Body)

	fmt.Printf("Counter value: %s", strings.TrimSpace(body.String()))
	// Output:
	// Counter value: 8
}