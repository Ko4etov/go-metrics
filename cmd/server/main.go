// Package main содержит точку входа для сервера сбора метрик.
//
// Программа запускает HTTP-сервер, который предоставляет API для:
// - Приема и хранения метрик от агентов
// - Отдачи сохраненных метрик по запросу
// - Работы с базой данных PostgreSQL (опционально)
// - Поддержки сжатия и хеширования данных
// - Аудита операций (опционально)
// - Профилирования (опционально)
//
// Конфигурация сервера загружается из:
//   - Флагов командной строки (приоритет 1)
//   - Переменных окружения (приоритет 2)
//   - Значений по умолчанию (приоритет 3)
//
// Поддерживаемые флаги командной строки:
//
//	-a: адрес сервера (пример: -a "localhost:8080")
//	-i: интервал сохранения метрик в секундах (пример: -i 300)
//	-f: путь к файлу хранения метрик (пример: -f "/tmp/metrics.json")
//	-r: восстанавливать метрики при старте (пример: -r true)
//	-d: адрес базы данных (пример: -d "postgres://user:pass@localhost:5432/db")
//	-k: ключ для хеширования (опционально)
//	--audit-file: файл для аудита (опционально)
//	--audit-url: URL для отправки аудита (опционально)
//	--profile: включить профилирование (опционально)
//
// Пример запуска:
//
//	go run cmd/server/main.go -a "localhost:8080" -i 300 -f "/tmp/metrics.json"
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ko4etov/go-metrics/internal/server"
	"github.com/Ko4etov/go-metrics/internal/server/config"
	"github.com/Ko4etov/go-metrics/internal/service/buildinfo"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// main является точкой входа для сервера сбора метрик.
// Функция выполняет:
//  1. Инициализацию конфигурации из флагов и переменных окружения
//  2. Проверку корректности конфигурации
//  3. Создание экземпляра сервера с полученной конфигурацией
//  4. Запуск HTTP-сервера для обработки запросов
//
// В случае ошибки при инициализации конфигурации программа завершается с panic.
func main() {
	info := buildinfo.New(buildVersion, buildDate, buildCommit)
	info.Print()

	ctx, stop := signal.NotifyContext(context.Background(), 
        syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
    defer stop()
	
	// Инициализация конфигурации сервера
	config, err := config.New()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

    // Создаем сервер с контекстом
    server := server.New(ctx, config)

    // Запускаем (блокируется до остановки)
    if err := server.Run(); err != nil {
        log.Fatal(err)
    }
}