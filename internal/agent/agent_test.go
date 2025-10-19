package agent

import (
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	agent := New(2*time.Second, 10*time.Second, "localhost:8080")

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
	agent := New(100*time.Millisecond, 1*time.Second, ":8080")

	// Запускаем агент в отдельной горутине
	go agent.Run()

	// Даем агенту поработать 300ms (примерно 3 сбора метрик)
	time.Sleep(300 * time.Millisecond)

	// Останавливаем агент
	agent.Stop()

	// Проверяем, что агент остановлен
	if agent.IsRunning() {
		t.Error("Agent should be stopped after Stop() call")
	}

	// Проверяем, что метрики собирались (должно быть минимум 2 сбора за 300ms)
	count := agent.collector.PollCount()
	if count < 2 {
		t.Errorf("Expected at least 2 collections, got %d", count)
	}
	
	t.Logf("Completed %d poll cycles", count)
}