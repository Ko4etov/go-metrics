package metricssenderfactory

import (
	"fmt"

	"github.com/Ko4etov/go-metrics/internal/agent/config"
	"github.com/Ko4etov/go-metrics/internal/agent/interfaces"
	grpcsender "github.com/Ko4etov/go-metrics/internal/agent/service/grpc_sender"
	metricssender "github.com/Ko4etov/go-metrics/internal/agent/service/metrics_sender"
)

// NewSender создает отправитель в зависимости от конфигурации.
func NewSender(cfg *config.AgentConfig) (interfaces.MetricsSender, error) {
	if cfg.UseGRPC {
		if (cfg.GRPCAddress == "") {
			return nil, fmt.Errorf("failed to create gRPC sender GRPCAddress is empty")
		}

		sender, err := grpcsender.New(cfg.GRPCAddress, cfg.HashKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC sender: %w", err)
		}
		
		return sender, nil
	}
	
	return metricssender.New(cfg.Address, cfg.HashKey, cfg.RateLimit, cfg.CryptoKey), nil
}