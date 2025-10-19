package main

import (
	"time"

	"github.com/Ko4etov/go-metrics/internal/agent"
	"github.com/Ko4etov/go-metrics/internal/agent/config"
)

func main() {
	agentConfig := config.New()

	// Создание агента
	agent.New(
		time.Duration(agentConfig.PollInterval)*time.Second,
		time.Duration(agentConfig.ReportInterval)*time.Second,
		agentConfig.AgentAddress,
	).Run()
}
