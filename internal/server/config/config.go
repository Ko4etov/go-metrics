package config

type ServerConfig struct {
	ServerAddress string
}

func New() *ServerConfig {
	parseServerParameters()

	return &ServerConfig{
		ServerAddress: address,
	}
}