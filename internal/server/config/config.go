// Package config предоставляет конфигурацию для сервера сбора метрик.
package config

import (
	"errors"
	"fmt"
	"net"

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
	CryptoKey              string        // директория для сохранения профилей
	TrustedNet             *net.IPNet    // доверенная подсеть (CIDR) для проверки IP
}

// New создает новую конфигурацию сервера.
func New() (*ServerConfig, error) {
	var pool *pgxpool.Pool

	if err := logger.Initialize("info"); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLogerInitialization, err)
	}

	serverParameters, err := parseServerParameters()
	if err != nil {
		return nil, err
	}

	_, trustedNet, parseCidrErr := net.ParseCIDR(serverParameters.TrustedSubnet)
	if parseCidrErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidTrustedSubnetParameter, err)
	}

	if serverParameters.DBAddress != "" {
		if _, err := pgxpool.ParseConfig(serverParameters.DBAddress); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrParseDBConfig, err)
		}

		if err := db.RunMigrations(serverParameters.DBAddress); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrMigration, err)
		}

		pool, err = db.NewDBConnection(serverParameters.DBAddress)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrDBConnection, err)
		}
	}

	return &ServerConfig{
		ServerAddress:          serverParameters.Address,
		StoreMetricsInterval:   serverParameters.StoreMetricsInterval,
		FileStorageMetricsPath: serverParameters.FileStorageMetricsPath,
		RestoreMetrics:         serverParameters.RestoreMetrics,
		ConnectionPool:         pool,
		HashKey:                serverParameters.HashKey,
		AuditFile:              serverParameters.AuditFile,
		AuditURL:               serverParameters.AuditURL,
		ProfilingEnable:        serverParameters.ProfilingEnable,
		ProfileServerAddress:   serverParameters.ProfileServerAddress,
		ProfilingDir:           serverParameters.ProfilingDir,
		CryptoKey:              serverParameters.CryptoKey,
		TrustedNet:             trustedNet,
	}, nil
}

var (
	ErrInvalidTrustedSubnetParameter = errors.New("invalid trusted subnet parameter")
	ErrMigration                     = errors.New("migration error")
	ErrParseDBConfig                 = errors.New("parse db config error")
	ErrDBConnection                  = errors.New("db connection error")
	ErrLogerInitialization           = errors.New("logger initialization error")
)
