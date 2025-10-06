package agent

import (
	"log"
	"time"

	"github.com/Ko4etov/go-metrics/internal/repository/collector"
	"github.com/Ko4etov/go-metrics/internal/service"
)

// Agent представляет основной агент сбора и отправки метрик
type Agent struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	serverAddress  string
	collector      collector.MetricsCollector
	sender         service.MetricsSenderInterface
}

// NewAgent создает новый экземпляр агента
func NewAgent(pollInterval, reportInterval time.Duration, serverAddress string) *Agent {
	collector := collector.New()
	sender := service.NewMetricsSenderService(serverAddress)

	return &Agent{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		serverAddress:  serverAddress,
		collector:      collector,
		sender:         sender,
	}
}

// Run запускает агент
func (a *Agent) Run() {
	for {
		// Запускаем сбор метрик
		a.pollMetrics()
		// Запускаем отправку метрик
		a.reportMetrics()
	}
}

// pollMetrics собирает метрики с заданным интервалом
func (a *Agent) pollMetrics() {
	log.Printf("pollMetrics")
	a.collector.Collect()
	time.Sleep(a.pollInterval)
}

// reportMetrics отправляет метрики на сервер с заданным интервалом
func (a *Agent) reportMetrics() {
	log.Printf("reportMetrics")
	metrics := a.collector.Metrics()
	a.sender.SendMetrics(metrics)
	time.Sleep(a.reportInterval)
}
