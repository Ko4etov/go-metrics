package agent

import (
	"sync"
	"time"

	"github.com/Ko4etov/go-metrics/internal/repository"
	"github.com/Ko4etov/go-metrics/internal/service"
)

// Agent представляет основной агент сбора и отправки метрик
type Agent struct {
    pollInterval   time.Duration
    reportInterval time.Duration
    serverAddress  string
    collector      repository.CollectorInterface
    sender         service.MetricsSenderInterface
    wg             sync.WaitGroup
    stopChan       chan struct{}
}

// NewAgent создает новый экземпляр агента
func NewAgent(pollInterval, reportInterval time.Duration, serverAddress string) *Agent {
    collector := repository.NewCollector()
    sender := service.NewMetricsSenderService(serverAddress)

    return &Agent{
        pollInterval:   pollInterval,
        reportInterval: reportInterval,
        serverAddress:  serverAddress,
        collector:      collector,
        sender:         sender,
        stopChan:       make(chan struct{}),
    }
}

// Run запускает агент
func (a *Agent) Run() {
    // Запускаем сбор метрик
    a.wg.Add(1)
    go a.pollMetrics()

    // Запускаем отправку метрик
    a.wg.Add(1)
    go a.reportMetrics()

    // Ждем сигнала остановки
    a.wg.Wait()
}

// Stop останавливает агент
func (a *Agent) Stop() {
    close(a.stopChan)
    a.wg.Wait()
}

// pollMetrics собирает метрики с заданным интервалом
func (a *Agent) pollMetrics() {
    defer a.wg.Done()

    ticker := time.NewTicker(a.pollInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            a.collector.Collect()
        case <-a.stopChan:
            return
        }
    }
}

// reportMetrics отправляет метрики на сервер с заданным интервалом
func (a *Agent) reportMetrics() {
    defer a.wg.Done()

    ticker := time.NewTicker(a.reportInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            metrics := a.collector.GetMetrics()
            a.sender.SendMetrics(metrics)
        case <-a.stopChan:
            return
        }
    }
}