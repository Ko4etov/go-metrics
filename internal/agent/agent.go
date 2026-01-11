package agent

import (
	"context"
	"sync"
	"time"

	"github.com/Ko4etov/go-metrics/internal/agent/config"
	"github.com/Ko4etov/go-metrics/internal/agent/interfaces"
	"github.com/Ko4etov/go-metrics/internal/agent/repository/collector"
	metricssender "github.com/Ko4etov/go-metrics/internal/agent/service/metrics_sender"
)

type Agent struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	serverAddress  string
	collector      interfaces.Collector
	sender         interfaces.MetricsSender
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	isRunning      bool
	mu             sync.RWMutex
}

func New(config *config.AgentConfig) *Agent {
	collector := collector.New()
	sender := metricssender.New(config.Address, config.HashKey, config.RateLimit)
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		pollInterval:   config.PollInterval,
		reportInterval: config.ReportInterval,
		serverAddress:  config.Address,
		collector:      collector,
		sender:         sender,
		ctx:            ctx,
		cancel:         cancel,
		isRunning:      false,
	}
}

func (a *Agent) Run() {
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return
	}
	a.isRunning = true
	a.mu.Unlock()

	a.wg.Add(2)
	go a.runPolling()
	go a.runReporting()

	a.wg.Wait()

	a.mu.Lock()
	a.isRunning = false
	a.mu.Unlock()
}

func (a *Agent) runPolling() {
	defer a.wg.Done()

	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.pollMetrics()
		}
	}
}

func (a *Agent) runReporting() {
	defer a.wg.Done()

	ticker := time.NewTicker(a.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.reportMetrics()
		}
	}
}

func (a *Agent) Stop() {
	a.mu.RLock()
	if !a.isRunning {
		a.mu.RUnlock()
		return
	}
	a.mu.RUnlock()

	a.cancel()

	if sender, ok := a.sender.(interface{ Stop() }); ok {
		sender.Stop()
	}

	a.wg.Wait()
}

func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isRunning
}

func (a *Agent) pollMetrics() {
	select {
	case <-a.ctx.Done():
		return
	default:
		a.collector.Collect()
	}
}

func (a *Agent) reportMetrics() {
	select {
	case <-a.ctx.Done():
		return
	default:
		metrics := a.collector.Metrics()
		a.sender.SendMetrics(metrics)
		a.collector.PollCountReset()
	}
}
