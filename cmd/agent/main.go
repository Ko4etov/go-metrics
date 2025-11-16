package main

import (
	"github.com/Ko4etov/go-metrics/internal/agent"
	"github.com/Ko4etov/go-metrics/internal/agent/config"
)

func main() {
	agentConfig := config.New()

	// Создание агента
	agent.New(agentConfig).Run()
}
