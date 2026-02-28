// Package factory предоставляет фабрику для создания сервера.
package serverfactory

import (
	"context"
	"fmt"
	"net"

	"github.com/Ko4etov/go-metrics/internal/server/config"
	grpcserver "github.com/Ko4etov/go-metrics/internal/server/server/grpc_server"
	httpserver "github.com/Ko4etov/go-metrics/internal/server/server/http_server"
	"google.golang.org/grpc"
)

// Server определяет интерфейс сервера.
type Server interface {
	// Start запускает сервер (блокирующий вызов)
	Start() error
	// Stop останавливает сервер
	Stop(ctx context.Context) error
}

// GRPCServer реализует gRPC сервер.
type GRPCServer struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer создает сервер на основе конфигурации.
func NewServer(ctx context.Context, cfg *config.ServerConfig) (Server, error) {
	if (cfg.UseGRPC) {
		if cfg.GRPCAddress != "" {
			return grpcserver.New(ctx, cfg)
		}
	}
	
	if cfg.ServerAddress != "" {
		return  httpserver.New(ctx, cfg)
	}
	
	return nil, fmt.Errorf("no server address configured")
}