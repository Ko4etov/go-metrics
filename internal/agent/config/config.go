package config

type AgentConfig struct {
	Address string
	PollInterval int
	ReportInterval int
}

func New() *AgentConfig {
	parseAgentParameters()

	return &AgentConfig{
		Address: address,
		PollInterval: pollInterval,
		ReportInterval: reportInterval,
	}
}