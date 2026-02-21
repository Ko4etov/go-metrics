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
	"os"
	"os/signal"
	"syscall"

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
	// Инициализация конфигурации агента
	agentConfig := config.New()

	// Создание и запуск агента
	agent := agent.New(agentConfig)

	go agent.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-quit

	agent.Stop()
}