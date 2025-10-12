package agent

import (
	"context"
	"sync"
	"time"

	"github.com/Ko4etov/go-metrics/internal/agent/interfaces"
	"github.com/Ko4etov/go-metrics/internal/agent/repository/collector"
	metricssender "github.com/Ko4etov/go-metrics/internal/agent/service/metrics_sender"
	pollcounter "github.com/Ko4etov/go-metrics/internal/agent/service/poll_counter"
)

// Agent представляет основной агент сбора и отправки метрик
type Agent struct {
	pollInterval       time.Duration
	reportInterval     time.Duration
	serverAddress      string
	collector          interfaces.Collector
	sender             interfaces.MetricsSender
	pollMetricsCounter *pollcounter.PollCounter
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
	isRunning          bool
	mu                 sync.RWMutex
}

// NewAgent создает новый экземпляр агента
func New(pollInterval time.Duration, reportInterval time.Duration, serverAddress string) *Agent {
	pollMetricsCounter := pollcounter.New("")
	collector := collector.New(pollMetricsCounter)
	sender := metricssender.New(serverAddress)
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		pollInterval:       pollInterval,
		reportInterval:     reportInterval,
		serverAddress:      serverAddress,
		collector:          collector,
		sender:             sender,
		pollMetricsCounter: pollMetricsCounter,
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
			a.pollMetricsCounter.Increment()
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
			a.pollMetricsCounter.Reset()
	}
}
