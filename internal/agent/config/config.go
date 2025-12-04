package config

import "time"

type AgentConfig struct {
	Address string
	PollInterval time.Duration
	ReportInterval time.Duration
	HashKey string
	RateLimit int
}

func New() *AgentConfig {
	parameters := parseAgentParameters()

	return &AgentConfig{
		Address: parameters.Address,
		PollInterval: time.Duration(parameters.PollInterval)*time.Second,
		ReportInterval: time.Duration(parameters.ReportInterval)*time.Second,
		HashKey: parameters.HashKey,
		RateLimit: parameters.RateLimit,
	}
}