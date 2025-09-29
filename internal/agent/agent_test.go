package agent

import (
	"sync"
	"testing"
	"time"

	"github.com/Ko4etov/go-metrics/internal/repository"
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
    mockCollector := &repository.MockCollector{}
    mockSender := &service.MockMetricsSenderService{}
    
    agent := &Agent{
        pollInterval:   100 * time.Millisecond,
        reportInterval: 1 * time.Second,
        collector:      mockCollector,
        sender:         mockSender,
        stopChan:       make(chan struct{}),
    }
    
    var wg sync.WaitGroup
    wg.Add(1)
    
    // Запускаем сбор метрик на короткое время
    go func() {
        defer wg.Done()
        agent.pollMetrics()
    }()
    
    // Даем время на несколько сборов
    time.Sleep(350 * time.Millisecond)
    close(agent.stopChan)
    wg.Wait()
    
    if mockCollector.CollectCount < 3 {
        t.Errorf("Expected at least 3 collections, got %d", mockCollector.CollectCount)
    }
}

func TestAgent_ReportMetrics(t *testing.T) {
    mockCollector := &repository.MockCollector{}
    mockSender := &service.MockMetricsSenderService{}
    
    agent := &Agent{
        pollInterval:   1 * time.Second,
        reportInterval: 100 * time.Millisecond,
        collector:      mockCollector,
        sender:         mockSender,
        stopChan:       make(chan struct{}),
    }
    
    var wg sync.WaitGroup
    wg.Add(1)
    
    // Запускаем отправку метрик на короткое время
    go func() {
        defer wg.Done()
        agent.reportMetrics()
    }()
    
    // Даем время на несколько отправок
    time.Sleep(350 * time.Millisecond)
    close(agent.stopChan)
    wg.Wait()
    
    if mockSender.SendCount < 3 {
        t.Errorf("Expected at least 3 sends, got %d", mockSender.SendCount)
    }
}

func TestAgent_Stop(t *testing.T) {
    agent := NewAgent(100*time.Millisecond, 100*time.Millisecond, "localhost:8080")
    
    // Запускаем агент в отдельной горутине
    go agent.Run()
    
    // Даем ему поработать немного
    time.Sleep(200 * time.Millisecond)
    
    // Останавливаем
    agent.Stop()
    
    // Проверяем, что горутины действительно остановились
    // (этот тест в основном проверяет, что нет паники при остановке)
}

func TestAgent_Integration(t *testing.T) {
    agent := NewAgent(50*time.Millisecond, 100*time.Millisecond, "localhost:8080")
    
    // Заменяем sender на mock для тестирования
    mockSender := &service.MockMetricsSenderService{}
    agent.sender = mockSender
    
    // Запускаем агент на короткое время
    go agent.Run()
    time.Sleep(250 * time.Millisecond)
    agent.Stop()
    
    // Проверяем, что метрики собирались и отправлялись
    pollCount := agent.collector.GetPollCount()
    if pollCount < 4 {
        t.Errorf("Expected at least 4 polls, got %d", pollCount)
    }
    
    if mockSender.SendCount < 2 {
        t.Errorf("Expected at least 2 sends, got %d", mockSender.SendCount)
    }
}