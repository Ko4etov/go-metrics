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

// Agent представляет основной агент сбора и отправки метрик
type Agent struct {
	pollInterval       time.Duration
	reportInterval     time.Duration
	serverAddress      string
	collector          interfaces.Collector
	sender             interfaces.MetricsSender
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
	isRunning          bool
	mu                 sync.RWMutex
}

// NewAgent создает новый экземпляр агента
func New(config *config.AgentConfig) *Agent {
	collector := collector.New()
	sender := metricssender.New(config.Address, config.HashKey)
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		pollInterval:       config.PollInterval,
		reportInterval:     config.ReportInterval,
		serverAddress:      config.Address,
		collector:          collector,
		sender:             sender,
		ctx:                ctx,
		cancel:             cancel,
		isRunning:          false,
	}
}

// Run запускает агент
func (a *Agent) Run() {
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return
	}
	a.isRunning = true
	a.mu.Unlock()

	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)

	defer func() {
		pollTicker.Stop()
		reportTicker.Stop()

		a.mu.Lock()
		a.isRunning = false
		a.mu.Unlock()
	}()

	a.wg.Add(1)
	defer a.wg.Done()

	for {
		select {
			case <-a.ctx.Done():
				return
			case <-pollTicker.C:
				a.pollMetrics()
			case <-reportTicker.C:
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
	
	a.wg.Wait()
}

func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isRunning
}

// pollMetrics собирает метрики с заданным интервалом
func (a *Agent) pollMetrics() {
	select {
		case <-a.ctx.Done():
			return
		default:
			a.collector.Collect()
	}
}

// reportMetrics отправляет метрики на сервер с заданным интервалом
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
