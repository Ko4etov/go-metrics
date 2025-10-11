package config

type ServerConfig struct {
	ServerAddress string
}

func New() *ServerConfig {
	parseServerFlags()

	return &ServerConfig{
		ServerAddress: serverAddress,
	}
}