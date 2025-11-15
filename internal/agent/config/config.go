package config

type AgentConfig struct {
	Address string
	PollInterval int
	ReportInterval int
}

func New() *AgentConfig {
	parameters := parseAgentParameters()

	return &AgentConfig{
		Address: parameters.Address,
		PollInterval: parameters.PollInterval,
		ReportInterval: parameters.ReportInterval,
	}
}