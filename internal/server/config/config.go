package config

import (
	"fmt"

	"github.com/Ko4etov/go-metrics/internal/server/config/db"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerConfig struct {
	ServerAddress string
	StoreMetricsInterval int
	FileStorageMetricsPath string
	RestoreMetrics bool
	ConnectionPool *pgxpool.Pool
	HashKey string
}

func New() (*ServerConfig, error) {
	var poll *pgxpool.Pool

	if err := logger.Initialize("info"); err != nil {
        return nil, fmt.Errorf("logger initialization error: %s", err)
    }

	serverParameters := parseServerParameters()

	if serverParameters.DBAddress != "" {
		if _, err := pgxpool.ParseConfig(serverParameters.DBAddress); err == nil {
			poll, err = db.NewDBConnection(serverParameters.DBAddress)
			if (err != nil) {
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
		ServerAddress: serverParameters.Address,
		StoreMetricsInterval: serverParameters.StoreMetricsInterval,
		FileStorageMetricsPath: serverParameters.FileStorageMetricsPath,
		RestoreMetrics: serverParameters.RestoreMetrics,
		ConnectionPool: poll,
		HashKey: serverParameters.HashKey,
	}, nil
}