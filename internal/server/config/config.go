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
}

func New() (*ServerConfig, error) {
	var poll *pgxpool.Pool

	if err := logger.Initialize("info"); err != nil {
        return nil, fmt.Errorf("logger initialization error: %s", err)
    }

	parseServerParameters()

	if dbAddress != "" {
		if _, err := pgxpool.ParseConfig(dbAddress); err == nil {
			poll, err = db.NewDBConnection(dbAddress)
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
		ServerAddress: address,
		StoreMetricsInterval: storeMetricsInterval,
		FileStorageMetricsPath: fileStorageMetricsPath,
		RestoreMetrics: restoreMetrics,
		ConnectionPool: poll,
	}, nil
}