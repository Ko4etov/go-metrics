package config

type ServerConfig struct {
	ServerAddress string
	StoreMetricsInterval int
	FileStorageMetricsPath string
	RestoreMetrics bool
}

func New() *ServerConfig {
	parseServerParameters()

	return &ServerConfig{
		ServerAddress: address,
		StoreMetricsInterval: storeMetricsInterval,
		FileStorageMetricsPath: fileStorageMetricsPath,
		RestoreMetrics: restoreMetrics,
	}
}