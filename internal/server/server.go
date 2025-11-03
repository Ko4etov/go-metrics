package server

import (
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/server/config"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/Ko4etov/go-metrics/internal/server/router"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

type Server struct {
	config  *config.ServerConfig
}

func New(config *config.ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Run() {
	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics: s.config.RestoreMetrics,
		StoreMetricsInterval: s.config.StoreMetricsInterval,
		FileStorageMetricsPath: s.config.FileStorageMetricsPath,
	}
	metricsStorage := storage.New(storageConfig)
	serverRouter := router.New(metricsStorage)
	defer metricsStorage.Stop()

	if err := logger.Initialize("info"); err != nil {
        panic(err)
    }

	// Запускаем сервер
	err := http.ListenAndServe(s.config.ServerAddress, serverRouter)
	if err != nil {
		panic(err)
	}
}