package baseserver

import (
	"fmt"
	"os"

	"github.com/Ko4etov/go-metrics/internal/server/config"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/Ko4etov/go-metrics/internal/server/service/audit"
	"github.com/Ko4etov/go-metrics/internal/server/service/profiler"
)

// BaseServer содержит общие компоненты.
type BaseServer struct {
	Config       *config.ServerConfig
	Storage      *storage.MetricsStorage
	AuditSvc     *audit.AuditService
	FileAuditors []*audit.FileAuditor
}

func New(config *config.ServerConfig) *BaseServer {
	return &BaseServer{
		Config: config,
	}
}

// initComponents инициализирует общие компоненты.
func (b *BaseServer) Start() error {
	if b.Config.ProfilingEnable {
		if err := os.MkdirAll(b.Config.ProfilingDir, 0755); err != nil {
			return fmt.Errorf("failed to create profile directory: %w", err)
		}
		profiler.StartProfiling(b.Config.ProfileServerAddress)
	}

	storageConfig := &storage.MetricsStorageConfig{
		RestoreMetrics:         b.Config.RestoreMetrics,
		StoreMetricsInterval:   b.Config.StoreMetricsInterval,
		FileStorageMetricsPath: b.Config.FileStorageMetricsPath,
		ConnectionPool:         b.Config.ConnectionPool,
	}
	b.Storage = storage.New(storageConfig)

	if b.Config.AuditFile != "" || b.Config.AuditURL != "" {
		if err := b.initAudit(); err != nil {
			return err
		}
	}

	if b.Config.StoreMetricsInterval > 0 {
		b.Storage.StartPeriodicSave()
	}

	return nil
}

// initAudit инициализирует систему аудита.
func (b *BaseServer) initAudit() error {
	b.AuditSvc = audit.NewAuditService()

	if b.Config.AuditFile != "" {
		fileAuditor, err := audit.NewFileAuditor(b.Config.AuditFile)
		if err != nil {
			return fmt.Errorf("failed to create file auditor: %w", err)
		}
		b.AuditSvc.Subscribe(fileAuditor)
		b.FileAuditors = append(b.FileAuditors, fileAuditor)
	}

	if b.Config.AuditURL != "" {
		httpAuditor := audit.NewHTTPAuditor(b.Config.AuditURL)
		b.AuditSvc.Subscribe(httpAuditor)
	}

	return nil
}

// shutdown выполняет общую логику остановки.
func (b *BaseServer) Stop() error {
	if b.Config.StoreMetricsInterval > 0 {
		if err := b.Storage.StopPeriodicSave(); err != nil {
			return fmt.Errorf("error stopping periodic save: %v", err)
		}
	} else if b.Config.FileStorageMetricsPath != "" {
		if err := b.Storage.SaveToFile(); err != nil {
			return fmt.Errorf("failed to save metrics: %v", err)
		}
	}

	// Закрытие файловых аудиторов
	for _, fa := range b.FileAuditors {
		if err := fa.Close(); err != nil {
			return fmt.Errorf("failed to close file auditor: %v", err)
		}
	}

	// Закрытие соединения с БД
	if b.Config.ConnectionPool != nil {
		b.Config.ConnectionPool.Close()
	}

	return nil
}
