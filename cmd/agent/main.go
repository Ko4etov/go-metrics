// Package main содержит точку входа для агента сбора метрик.
//
// Программа запускает агента, который периодически:
// - Собирает метрики системы (использование CPU, памяти, runtime метрики Go)
// - Отправляет метрики на сервер с настраиваемыми интервалами
// - Поддерживает rate limiting и хеширование данных
//
// Конфигурация агента загружается из:
//   - Флагов командной строки (приоритет 1)
//   - Переменных окружения (приоритет 2)
//   - Значений по умолчанию (приоритет 3)
//
// Поддерживаемые флаги командной строки:
//
//	-a: адрес сервера (пример: -a "localhost:8080")
//	-r: интервал отправки метрик в секундах (пример: -r 10)
//	-p: интервал сбора метрик в секундах (пример: -p 2)
//	-k: ключ для хеширования (опционально)
//	-l: лимит одновременных запросов (пример: -l 3)
//
// Пример запуска:
//
//	go run cmd/agent/main.go -a "localhost:8080" -r 10 -p 2
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ko4etov/go-metrics/internal/agent"
	"github.com/Ko4etov/go-metrics/internal/agent/config"
	"github.com/Ko4etov/go-metrics/internal/service/buildinfo"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// main является точкой входа для агента сбора метрик.
// Функция выполняет:
//  1. Инициализацию конфигурации из флагов и переменных окружения
//  2. Создание экземпляра агента с полученной конфигурацией
//  3. Запуск основного цикла работы агента
func main() {
	info := buildinfo.New(buildVersion, buildDate, buildCommit)
	info.Print()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	agentConfig, err := config.New()
	if (err != nil) {
		log.Fatalf("Config generation error: %v", err)
		return
	}

	agent, err := agent.New(ctx, agentConfig)
	if (err != nil) {
		log.Fatalf("Agent error: %v", err)
		return
	}

	if err := agent.Run(); err != nil {
		log.Fatalf("Agent error: %v", err)
		return
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := agent.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}