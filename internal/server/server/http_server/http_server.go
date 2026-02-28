// Package server предоставляет сервер для системы сбора метрик.
package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/server/config"
	"github.com/Ko4etov/go-metrics/internal/server/router"
	baseServer "github.com/Ko4etov/go-metrics/internal/server/server/base_server"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

type HTTPServer struct {
	*baseServer.BaseServer
	httpServer *http.Server
	Ctx        context.Context
	Cancel     context.CancelFunc
}

// New создает новый экземпляр сервера.
func New(ctx context.Context, config *config.ServerConfig) (*HTTPServer, error) {
	ctx, cancel := context.WithCancel(ctx)

	baseServer := baseServer.New(config)

	server := &HTTPServer{
		BaseServer: baseServer,
		Ctx: ctx,
		Cancel: cancel,
	}

	return server, nil
}

// Run запускает HTTP-сервер.
func (s *HTTPServer) Start() error {
    if err := s.BaseServer.Start(); err != nil {
        return err
    }

    routerConfig := &router.RouteConfig{
        Storage:    s.BaseServer.Storage,
        Pgx:        s.BaseServer.Config.ConnectionPool,
        HashKey:    s.BaseServer.Config.HashKey,
        AuditSvc:   s.BaseServer.AuditSvc,
        CryptoKey:  s.BaseServer.Config.CryptoKey,
        TrustedNet: s.BaseServer.Config.TrustedNet,
    }

    serverRouter, err := router.New(routerConfig)
    if err != nil {
        return fmt.Errorf("router initialization failed: %w", err)
    }

    s.httpServer = &http.Server{
        Addr:    s.BaseServer.Config.ServerAddress,
        Handler: serverRouter,
    }

	errorsChan := make(chan error, 1)

    go func() {
        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errorsChan <- fmt.Errorf("http server error: %v", err)
        }
    }()

	select {
	case err := <-errorsChan:
		return err
	case <-s.Ctx.Done():
		logger.Logger.Info("Received shutdown signal")
	}

    return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {
    if s.Cancel != nil {
        s.Cancel()
    }

	errorChan := make(chan error, 1)

    go func() {
        defer close(errorChan)

        if err := s.httpServer.Shutdown(context.Background()); err != nil {
			errorChan <- fmt.Errorf("http server shutdown error: %w", err)
			return
        }

        if err := s.BaseServer.Stop(); err != nil {
			errorChan <- fmt.Errorf("base server shutdown error: %w", err)
			return
        }
		errorChan <- nil
    }()

    select {
    case err := <-errorChan:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
