package agent

import (
	"testing"
	"time"

	"github.com/Ko4etov/go-metrics/internal/agent/config"
)

func TestNewAgent(t *testing.T) {
	config := &config.AgentConfig{
		ReportInterval: time.Duration(10) * time.Second,
		PollInterval:   time.Duration(2) * time.Second,
		Address:        ":8080",
		RateLimit:      1,
	}
	agent := New(config)

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
	config := &config.AgentConfig{
		PollInterval:   50 * time.Millisecond,
		ReportInterval: 500 * time.Millisecond,
		Address:        ":8080",
		RateLimit:      1,
	}
	agent := New(config)

	// Запускаем агент в отдельной горутине
	go agent.Run()

	// Даем агенту поработать 300ms (примерно 3 сбора метрик)
	time.Sleep(350 * time.Millisecond)

	// Останавливаем агент
	agent.Stop()

	time.Sleep(50 * time.Millisecond)

	// Проверяем, что агент остановлен
	if agent.IsRunning() {
		t.Error("Agent should be stopped after Stop() call")
	}
}
