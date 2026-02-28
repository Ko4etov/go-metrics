package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/Ko4etov/go-metrics/internal/proto"
	"github.com/Ko4etov/go-metrics/internal/server/config"
	grpcstorage "github.com/Ko4etov/go-metrics/internal/server/repository/grpc_storage"
	baseServer "github.com/Ko4etov/go-metrics/internal/server/server/base_server"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// Server реализует gRPC сервер, удовлетворяющий интерфейсу interfaces.Server.
type GRPCServer struct {
	baseServer.BaseServer
	grpcServer *grpc.Server
	listener   net.Listener
	Ctx        context.Context
	Cancel     context.CancelFunc
}

func New(ctx context.Context, config *config.ServerConfig) (*GRPCServer, error) {
	baseServer := baseServer.New(config)

	listener, err := net.Listen("tcp", config.GRPCAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(IPInterceptor(config.TrustedNet)),
	)

	ctx, cancel := context.WithCancel(ctx)

	return &GRPCServer{
		BaseServer: *baseServer,
		grpcServer: grpcServer,
		listener:   listener,
		Ctx:        ctx,
		Cancel:     cancel,
	}, nil
}

func (s *GRPCServer) Start() error {
	if err := s.BaseServer.Start(); err != nil {
		return err
	}

	grpcMetricsStorage := grpcstorage.NewGrpcMetricsStorage(s.Storage)

	proto.RegisterMetricsServer(s.grpcServer, grpcMetricsStorage)

	errorsChan := make(chan error, 1)

	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			errorsChan <- fmt.Errorf("gRPC server error: %w", err)
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

func (s *GRPCServer) Stop(ctx context.Context) error {
    if s.Cancel != nil {
        s.Cancel()
    }

	errorChan := make(chan error, 1)
    
    go func() {
		defer close(errorChan)

        s.grpcServer.GracefulStop()
        
        if err := s.BaseServer.Stop(); err != nil {
			errorChan <- fmt.Errorf("Base server shutdown error: %w", err)
			return
        }
		errorChan <- nil
    }()

    select {
    case err := <-errorChan:
        return err
    case <-ctx.Done():
        s.grpcServer.Stop()
        return ctx.Err()
    }
}

func IPInterceptor(trustedSubnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var clientIP string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ips := md.Get("x-real-ip"); len(ips) > 0 {
				clientIP = ips[0]
			}
		}

		if clientIP == "" {
			if p, ok := peer.FromContext(ctx); ok {
				clientIP, _, _ = net.SplitHostPort(p.Addr.String())
			}
		}

		logger.Logger.Infof("🔍 Client IP: %s", clientIP)

		if clientIP == "" {
			logger.Logger.Warn("Could not determine client IP")
			return nil, status.Error(codes.PermissionDenied, "IP address required")
		}

		ip := net.ParseIP(clientIP)
		if ip == nil {
			logger.Logger.Warnf("Invalid IP format: %s", clientIP)
			return nil, status.Error(codes.PermissionDenied, "invalid IP format")
		}

		if !trustedSubnet.Contains(ip) {
			logger.Logger.Warnf("IP %s not in trusted subnet %s", clientIP, trustedSubnet)
			return nil, status.Error(codes.PermissionDenied, "IP not allowed")
		}

		return handler(ctx, req)
	}
}
