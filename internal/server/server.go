package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Ko4etov/go-metrics/internal/server/config"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/Ko4etov/go-metrics/internal/server/router"
	"github.com/Ko4etov/go-metrics/internal/server/service/audit"
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
		ConnectionPool: s.config.ConnectionPool,
	}

	metricsStorage := storage.New(storageConfig)

	var auditSvc *audit.AuditService
	
	// Проверяем, нужно ли включать аудит
	if s.config.AuditFile != "" || s.config.AuditUrl != "" {
		auditSvc = audit.NewAuditService()
		
		if s.config.AuditFile != "" {
			fileAuditor, err := audit.NewFileAuditor(s.config.AuditFile)
			if err != nil {
				fmt.Printf("Failed to create file auditor: %v\n", err)
				os.Exit(1)
			}
			defer fileAuditor.Close()
			auditSvc.Subscribe(fileAuditor)
		}
		
		if s.config.AuditUrl != "" {
			httpAuditor := audit.NewHTTPAuditor(s.config.AuditUrl)
			auditSvc.Subscribe(httpAuditor)
		}
	}

	routerConfig := &router.RouteConfig{
		Storage: metricsStorage,
		Pgx: s.config.ConnectionPool,
		HashKey: s.config.HashKey,
		AuditSvc: auditSvc,
	}
	serverRouter := router.New(routerConfig)

	if s.config.StoreMetricsInterval > 0 {
		metricsStorage.StartPeriodicSave()
		defer metricsStorage.StopPeriodicSave()
	}

	// Запускаем сервер
	err := http.ListenAndServe(s.config.ServerAddress, serverRouter)
	if err != nil {
		panic(err)
	}
}