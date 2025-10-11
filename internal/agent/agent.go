package agent

import (
	"time"

	"github.com/Ko4etov/go-metrics/internal/agent/interfaces"
	"github.com/Ko4etov/go-metrics/internal/agent/repository/collector"
	"github.com/Ko4etov/go-metrics/internal/agent/service/metrics_sender"
	"github.com/Ko4etov/go-metrics/internal/agent/service/poll_metrics_counter"
)

// Agent представляет основной агент сбора и отправки метрик
type Agent struct {
	pollInterval       time.Duration
	reportInterval     time.Duration
	serverAddress      string
	collector          interfaces.Collector
	sender             interfaces.MetricsSender
	pollMetricsCounter *poll_metrics_counter.PollMetricsCounter
}

// NewAgent создает новый экземпляр агента
func NewAgent(pollInterval time.Duration, reportInterval time.Duration, serverAddress string) *Agent {
	pollMetricsCounter := poll_metrics_counter.New()
	collector := collector.New(pollMetricsCounter)
	sender := metrics_sender.New(serverAddress)

	return &Agent{
		pollInterval:       pollInterval,
		reportInterval:     reportInterval,
		serverAddress:      serverAddress,
		collector:          collector,
		sender:             sender,
		pollMetricsCounter: pollMetricsCounter,
	}
}

// Run запускает агент
func (a *Agent) Run() {
	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)

	defer reportTicker.Stop()
	defer pollTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			a.pollMetrics()
		case <-reportTicker.C:
			a.reportMetrics()
		}
	}
}

// pollMetrics собирает метрики с заданным интервалом
func (a *Agent) pollMetrics() {
	a.collector.Collect()
	a.pollMetricsCounter.Increment()
}

// reportMetrics отправляет метрики на сервер с заданным интервалом
func (a *Agent) reportMetrics() {
	metrics := a.collector.Metrics()
	a.sender.SendMetrics(metrics)
	a.pollMetricsCounter.Reset()
}
