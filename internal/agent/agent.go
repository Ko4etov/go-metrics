// Package agent реализует агента для сбора и отправки метрик.
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

// Agent реализует агента для сбора и отправки метрик.
type Agent struct {
	pollInterval   time.Duration // интервал сбора метрик
	reportInterval time.Duration // интервал отправки метрик
	serverAddress  string        // адрес сервера
	collector      interfaces.Collector // сборщик метрик
	sender         interfaces.MetricsSender // отправитель метрик
	ctx            context.Context // контекст для управления жизненным циклом
	cancel         context.CancelFunc // функция отмены контекста
	wg             sync.WaitGroup // группа ожидания для горутин
	isRunning      bool // флаг работы агента
	mu             sync.RWMutex // мьютекс для безопасного доступа
}

// New создает нового агента.
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

// Run запускает агента.
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

// runPolling запускает горутину для периодического сбора метрик.
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

// runReporting запускает горутину для периодической отправки метрик.
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

// Stop останавливает агента.
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

// IsRunning возвращает состояние агента.
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isRunning
}

// pollMetrics собирает метрики.
func (a *Agent) pollMetrics() {
	select {
	case <-a.ctx.Done():
		return
	default:
		a.collector.Collect()
	}
}

// reportMetrics отправляет метрики на сервер.
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