// Package config предоставляет конфигурацию для сервера сбора метрик.
package config

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ko4etov/go-metrics/internal/server/config/db"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

// ServerConfig содержит все параметры конфигурации сервера.
type ServerConfig struct {
	ServerAddress          string        // адрес сервера
	StoreMetricsInterval   int           // интервал сохранения метрик в секундах
	FileStorageMetricsPath string        // путь к файлу хранения метрик
	RestoreMetrics         bool          // восстанавливать ли метрики при старте
	ConnectionPool         *pgxpool.Pool // пул подключений к базе данных
	HashKey                string        // ключ для хеширования
	AuditFile              string        // файл для аудита
	AuditURL               string        // URL для отправки аудита
	ProfilingEnable        bool          // включить профилирование
	ProfileServerAddress   string        // адрес сервера профилирования
	ProfilingDir           string        // директория для сохранения профилей
}

// New создает новую конфигурацию сервера.
func New() (*ServerConfig, error) {
	var poll *pgxpool.Pool

	if err := logger.Initialize("info"); err != nil {
		return nil, fmt.Errorf("logger initialization error: %s", err)
	}

	serverParameters := parseServerParameters()

	if serverParameters.DBAddress != "" {
		if _, err := pgxpool.ParseConfig(serverParameters.DBAddress); err == nil {
			poll, err = db.NewDBConnection(serverParameters.DBAddress)
			if err != nil {
				return nil, fmt.Errorf("db initialization error: %v", err)
			}
		} else {
			return nil, fmt.Errorf("parse db config error: %v", err)
		}
	}

	if poll != nil {
		if err := db.RunMigrations(poll); err != nil {
			return nil, fmt.Errorf("migration error: %v", err)
		}
	}

	return &ServerConfig{
		ServerAddress:          serverParameters.Address,
		StoreMetricsInterval:   serverParameters.StoreMetricsInterval,
		FileStorageMetricsPath: serverParameters.FileStorageMetricsPath,
		RestoreMetrics:         serverParameters.RestoreMetrics,
		ConnectionPool:         poll,
		HashKey:                serverParameters.HashKey,
		AuditFile:              serverParameters.AuditFile,
		AuditURL:               serverParameters.AuditURL,
		ProfilingEnable:        serverParameters.ProfilingEnable,
		ProfileServerAddress:   serverParameters.ProfileServerAddress,
		ProfilingDir:           serverParameters.ProfilingDir,
	}, nil
}