package agent

import (
	"testing"
	"time"

	"github.com/Ko4etov/go-metrics/internal/repository/collector"
	"github.com/Ko4etov/go-metrics/internal/service"
)

func TestNewAgent(t *testing.T) {
	agent := NewAgent(2*time.Second, 10*time.Second, "localhost:8080")

	if agent == nil {
		t.Fatal("NewAgent() returned nil")
	}

	if agent.pollInterval != 2*time.Second {
		t.Errorf("Expected pollInterval 2s, got %v", agent.pollInterval)
	}

	if agent.reportInterval != 10*time.Second {
		t.Errorf("Expected reportInterval 10s, got %v", agent.reportInterval)
	}

	if agent.collector == nil {
		t.Error("Collector should be initialized")
	}

	if agent.sender == nil {
		t.Error("Sender should be initialized")
	}
}

func TestAgent_PollMetrics(t *testing.T) {
	mockCollector := &collector.CollectorMock{}
	mockSender := &service.MockMetricsSenderService{}

	agent := &Agent{
		pollInterval:   100 * time.Millisecond,
		reportInterval: 1 * time.Second,
		collector:      mockCollector,
		sender:         mockSender,
	}

	// Запускаем сбор метрик на короткое время
	agent.pollMetrics()

	if mockCollector.CollectCount < 1 {
		t.Errorf("Expected at least 1 collections, got %d", mockCollector.CollectCount)
	}
}

func TestAgent_ReportMetrics(t *testing.T) {
	mockCollector := &collector.CollectorMock{}
	mockSender := &service.MockMetricsSenderService{}

	agent := &Agent{
		pollInterval:   1 * time.Second,
		reportInterval: 100 * time.Millisecond,
		collector:      mockCollector,
		sender:         mockSender,
	}

	agent.reportMetrics()

	if mockSender.SendCount < 1 {
		t.Errorf("Expected at least 1 sends, got %d", mockSender.SendCount)
	}
}
