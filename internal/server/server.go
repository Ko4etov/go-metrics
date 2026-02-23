// Package server предоставляет сервер для системы сбора метрик.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Ko4etov/go-metrics/internal/server/config"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/Ko4etov/go-metrics/internal/server/router"
	"github.com/Ko4etov/go-metrics/internal/server/service/audit"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
	"github.com/Ko4etov/go-metrics/internal/server/service/profiler"
)

// Server представляет HTTP-сервер для системы метрик.
type Server struct {
	config       *config.ServerConfig // конфигурация сервера
	httpSrv      *http.Server         // HTTP-сервер
	storage      *storage.MetricsStorage
	auditSvc     *audit.AuditService
	fileAuditors []*audit.FileAuditor
	ctx          context.Context
	cancel       context.CancelFunc
}

// New создает новый экземпляр сервера.
func New(ctx context.Context, config *config.ServerConfig) *Server {
	serverCtx, cancel := context.WithCancel(ctx)
	
	return &Server{
		config: config,
		ctx:    serverCtx,
		cancel: cancel,
	}
}

// Run запускает HTTP-сервер.
func (s *Server) Run() error {
	if s.config.ProfilingEnable {
		if err := os.MkdirAll(s.config.ProfilingDir, 0755); err != nil {
			logger.Logger.Fatalf("failed to create profile directory: %v", err)
		}

		profiler.StartProfiling(s.config.ProfileServerAddress)

		profiler.SaveProfiling(s.config.ProfilingDir, 30*time.Second)
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:         s.config.RestoreMetrics,
		StoreMetricsInterval:   s.config.StoreMetricsInterval,
		FileStorageMetricsPath: s.config.FileStorageMetricsPath,
		ConnectionPool:         s.config.ConnectionPool,
	}

	s.storage = storage.New(storageConfig)

	if s.config.AuditFile != "" || s.config.AuditURL != "" {
		s.auditSvc = audit.NewAuditService()

		if s.config.AuditFile != "" {
			fileAuditor, err := audit.NewFileAuditor(s.config.AuditFile)
			if err != nil {
				return fmt.Errorf("failed to create file auditor: %v", err)
			}
			s.auditSvc.Subscribe(fileAuditor)
			s.fileAuditors = append(s.fileAuditors, fileAuditor)
		}

		if s.config.AuditURL != "" {
			httpAuditor := audit.NewHTTPAuditor(s.config.AuditURL)
			s.auditSvc.Subscribe(httpAuditor)
		}
	}

	routerConfig := &router.RouteConfig{
		Storage:   s.storage,
		Pgx:       s.config.ConnectionPool,
		HashKey:   s.config.HashKey,
		AuditSvc:  s.auditSvc,
		CryptoKey: s.config.CryptoKey,
	}
	serverRouter := router.New(routerConfig)

	s.httpSrv = &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: serverRouter,
	}

	if s.config.StoreMetricsInterval > 0 {
		s.storage.StartPeriodicSave()
	}

	serverErrors := make(chan error, 1)

	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-s.ctx.Done():
		logger.Logger.Info("Received shutdown signal")
	}

	// Graceful shutdown
	return s.Shutdown()
}

func (s *Server) Shutdown() error {
	logger.Logger.Info("Starting graceful shutdown...")

	// Контекст с таймаутом для операций завершения
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Остановка HTTP-сервера
	logger.Logger.Info("Shutting down HTTP server...")
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		logger.Logger.Errorf("HTTP server shutdown error: %v", err)
	}

	// Остановка периодического сохранения
	if s.config.StoreMetricsInterval > 0 {
		logger.Logger.Info("Stopping periodic save...")
		s.storage.StopPeriodicSave()
	} else if s.config.FileStorageMetricsPath != "" {
		logger.Logger.Info("Saving metrics to file...")
		if err := s.storage.SaveToFile(); err != nil {
			logger.Logger.Errorf("Failed to save metrics: %v", err)
		}
	}

	// Закрытие файловых аудиторов
	for _, fa := range s.fileAuditors {
		logger.Logger.Infof("Closing file auditor: %s", fa.Name())
		if err := fa.Close(); err != nil {
			logger.Logger.Errorf("Failed to close file auditor: %v", err)
		}
	}

	// Закрытие соединения с БД
	if s.config.ConnectionPool != nil {
		logger.Logger.Info("Closing database connection...")
		s.config.ConnectionPool.Close()
	}

	logger.Logger.Info("Server stopped gracefully")
	return nil
}
