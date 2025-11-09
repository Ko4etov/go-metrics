package config

import (
	"github.com/Ko4etov/go-metrics/internal/server/config/db"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerConfig struct {
	ServerAddress string
	StoreMetricsInterval int
	FileStorageMetricsPath string
	RestoreMetrics bool
	ConnectionPoll *pgxpool.Pool
}

func New() *ServerConfig {
	var poll *pgxpool.Pool

	if err := logger.Initialize("info"); err != nil {
        panic(err)
    }

	parseServerParameters()

	if _, err := pgxpool.ParseConfig(dbAddress); err == nil {
		poll = db.NewDBConnection(dbAddress)
	}


	return &ServerConfig{
		ServerAddress: address,
		StoreMetricsInterval: storeMetricsInterval,
		FileStorageMetricsPath: fileStorageMetricsPath,
		RestoreMetrics: restoreMetrics,
		ConnectionPoll: poll,
	}
}