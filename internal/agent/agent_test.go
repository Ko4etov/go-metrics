package agent

import (
	"context"
	"testing"
	"time"

	"github.com/Ko4etov/go-metrics/internal/agent/config"
)

func TestNewAgent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

	config := &config.AgentConfig{
		ReportInterval: time.Duration(10) * time.Second,
		PollInterval:   time.Duration(2) * time.Second,
		Address:        ":8080",
		RateLimit:      1,
	}

	agent := New(ctx, config)

	go func() {
		if err := agent.Run(); err != nil {
			t.Logf("Agent run finished: %v", err)
		}
	}()

	time.Sleep(350 * time.Millisecond)

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &config.AgentConfig{
		PollInterval:   50 * time.Millisecond,
		ReportInterval: 500 * time.Millisecond,
		Address:        "localhost:8080",
		RateLimit:      1,
		HashKey:        "test-key",
	}
	
	agent := New(ctx, cfg)

	go func() {
		if err := agent.Run(); err != nil {
			t.Logf("Agent run finished: %v", err)
		}
	}()

	time.Sleep(350 * time.Millisecond)

	if !agent.IsRunning() {
		t.Error("Agent should be running during work period")
	}

	cancel()

	time.Sleep(100 * time.Millisecond)

	if agent.IsRunning() {
		t.Error("Agent should be stopped after context cancel")
	}
}
