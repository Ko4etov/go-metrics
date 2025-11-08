package config

import (
	"github.com/Ko4etov/go-metrics/internal/server/config/db"
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
	parseServerParameters()

	poll := db.NewDBConnection(dbAddress)

	return &ServerConfig{
		ServerAddress: address,
		StoreMetricsInterval: storeMetricsInterval,
		FileStorageMetricsPath: fileStorageMetricsPath,
		RestoreMetrics: restoreMetrics,
		ConnectionPoll: poll,
	}
}