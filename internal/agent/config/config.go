package config

type AgentConfig struct {
	AgentAddress string
	PollInterval int
	ReportInterval int
}

func New() *AgentConfig {
	parseAgentFlags()

	return &AgentConfig{
		AgentAddress: agentAddress,
		PollInterval: pollInterval,
		ReportInterval: reportInterval,
	}
}