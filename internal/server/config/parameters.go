// Package config содержит параметры конфигурации сервера сбора метрик.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

const (
	address                = ":8080"        // Адрес сервера по умолчанию
	storeMetricsInterval   = 300            // Интервал сохранения метрик по умолчанию
	fileStorageMetricsPath = "metrics.json" // Путь к файлу метрик по умолчанию
	restoreMetrics         = true           // Восстанавливать метрики по умолчанию
	profilingEnable        = false          // Профилирование отключено по умолчанию
)

type ServerConfigFile struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	TrustedSubnet string `json:"trusted_subnet"`
}

// ServerParameters содержит все параметры конфигурации сервера.
type ServerParameters struct {
	Address                string // Адрес сервера
	StoreMetricsInterval   int    // Интервал сохранения метрик в секундах
	FileStorageMetricsPath string // Путь к файлу хранения метрик
	RestoreMetrics         bool   // Восстанавливать ли метрики при старте
	DBAddress              string // Адрес базы данных
	HashKey                string // Ключ для хеширования
	AuditFile              string // Файл для аудита
	AuditURL               string // URL для отправки аудита
	ProfilingEnable        bool   // Включить профилирование
	ProfileServerAddress   string // Адрес сервера профилирования
	ProfilingDir           string // Директория для сохранения профилей
	CryptoKey              string // Файл с крипто ключом
	TrustedSubnet          string // Доверенная подсеть (CIDR) для проверки IP
	UseGRPC                bool
	GRPCAddress            string
}

// parseServerParameters парсит параметры сервера из переменных окружения и флагов.
func parseServerParameters() (*ServerParameters, error) {
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info(".env file not loaded: %v", err)
	}

	addressParameter := addressParameter()
	storeMetricsIntervalParameter := storeMetricsIntervalParameter()
	fileStorageMetricsPathParameter := fileStorageMetricsPathParameter()
	restoreMetricsParameter := restoreMetricsParameter()
	dbAddressParameter := dbAddressParameter()
	hashKeyParameter, err := hashKeyParameter()
	if err != nil {
		return nil, err
	}
	auditFileParameter := auditFileParameter()
	AuditURLParameter := auditURLParameter()
	profilingEnableParameter := profilingEnableParameter()
	profileServerParameter := profileServerAddressParameter()
	profileDirParameter := profileDirParameter()
	cryptoKeyParameter := cryptoKeyParameter()
	configFileParameter := configFileParameter()
	trustedSubnetParameter := trustedSubnetParameter()
	useGRPCParameter := useGRPCParameter()
	grpcAddressParameter := grpcAddressParameter()

	flag.Parse()

	parameters := &ServerParameters{
		Address:                addressParameter,
		StoreMetricsInterval:   storeMetricsIntervalParameter,
		FileStorageMetricsPath: fileStorageMetricsPathParameter,
		RestoreMetrics:         restoreMetricsParameter,
		DBAddress:              dbAddressParameter,
		HashKey:                hashKeyParameter,
		AuditFile:              auditFileParameter,
		AuditURL:               AuditURLParameter,
		ProfilingEnable:        profilingEnableParameter,
		ProfileServerAddress:   profileServerParameter,
		ProfilingDir:           profileDirParameter,
		CryptoKey:              cryptoKeyParameter,
		TrustedSubnet:          trustedSubnetParameter,
		UseGRPC:                useGRPCParameter,
		GRPCAddress:            grpcAddressParameter,
	}

	if configFileParameter == "" {
		return parameters, nil
	}

	if err := loadFromConfigFile(parameters, configFileParameter); err != nil {
		logger.Logger.Warn("Failed to load config file: %v", err)
	}

	return parameters, nil
}

// loadFromConfigFile загружает параметры из JSON файла
func loadFromConfigFile(parameters *ServerParameters, configFilePath string) error {

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var configFile ServerConfigFile
	if err := json.Unmarshal(data, &configFile); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	if configFile.Address != "" && parameters.Address == address {
		parameters.Address = configFile.Address
	}

	if configFile.StoreInterval != "" && parameters.StoreMetricsInterval == storeMetricsInterval {
		if interval, err := time.ParseDuration(configFile.StoreInterval); err == nil {
			parameters.StoreMetricsInterval = int(interval.Seconds())
		}
	}

	if configFile.StoreFile != "" && parameters.FileStorageMetricsPath == fileStorageMetricsPath {
		parameters.FileStorageMetricsPath = configFile.StoreFile
	}

	defaultRestore := restoreMetrics
	if parameters.RestoreMetrics == defaultRestore {
		parameters.RestoreMetrics = configFile.Restore
	}

	if configFile.DatabaseDSN != "" && parameters.DBAddress == "" {
		parameters.DBAddress = configFile.DatabaseDSN
	}

	if configFile.CryptoKey != "" && parameters.CryptoKey == "" {
		parameters.CryptoKey = configFile.CryptoKey
	}

	return nil
}

// hashKeyParameter возвращает ключ для хеширования из переменных окружения или флагов.
func hashKeyParameter() (string, error) {
	env, ok := os.LookupEnv("KEY")

	if !ok {
		return "", fmt.Errorf("specify KEY parameter in ENV file")
	}

	flag.StringVar(&env, "k", env, "Hash key")

	return env, nil
}

// dbAddressParameter возвращает адрес базы данных из переменных окружения или флагов.
func dbAddressParameter() string {
	dbAddress := ""

	if env, ok := os.LookupEnv("DATABASE_DSN"); ok {
		dbAddress = env
	}

	flag.StringVar(&dbAddress, "d", dbAddress, "DB address")

	return dbAddress
}

// addressParameter возвращает адрес сервера из переменных окружения или флагов.
func addressParameter() string {
	address := address

	if env, ok := os.LookupEnv("ADDRESS"); ok {
		address = env
	}
	flag.StringVar(&address, "a", address, "Server address")

	return address
}

// storeMetricsIntervalParameter возвращает интервал сохранения метрик.
func storeMetricsIntervalParameter() int {
	storeMetricsInterval := storeMetricsInterval

	if storeMetricsIntervalEnv, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		if val, err := strconv.Atoi(storeMetricsIntervalEnv); err == nil {
			storeMetricsInterval = val
		}
	}
	flag.IntVar(&storeMetricsInterval, "i", storeMetricsInterval, "store metrics interval in seconds")

	return storeMetricsInterval
}

// fileStorageMetricsPathParameter возвращает путь к файлу хранения метрик.
func fileStorageMetricsPathParameter() string {
	fileStorageMetricsPath := fileStorageMetricsPath

	if fileStorageMetricsPathEnv, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		return fileStorageMetricsPathEnv
	}

	flag.StringVar(&fileStorageMetricsPath, "f", fileStorageMetricsPath, "file storage path")

	return fileStorageMetricsPath
}

// restoreMetricsParameter возвращает флаг восстановления метрик.
func restoreMetricsParameter() bool {
	restoreMetrics := restoreMetrics

	restoreMetricsEnv, ok := os.LookupEnv("RESTORE")

	if ok {
		if val, err := strconv.ParseBool(restoreMetricsEnv); err == nil {
			restoreMetrics = val
		}
	}

	flag.BoolVar(&restoreMetrics, "r", restoreMetrics, "restore metrics")

	return restoreMetrics
}

// auditFileParameter возвращает путь к файлу аудита.
func auditFileParameter() string {
	auditFile := ""

	if auditFileEnv, ok := os.LookupEnv("AUDIT_FILE"); ok {
		return auditFileEnv
	}

	flag.StringVar(&auditFile, "audit-file", auditFile, "Path to audit log file")

	return auditFile
}

// auditURLParameter возвращает URL для отправки аудита.
func auditURLParameter() string {
	AuditURL := ""

	if AuditURLEnv, ok := os.LookupEnv("AUDIT_URL"); ok {
		return AuditURLEnv
	}

	flag.StringVar(&AuditURL, "audit-url", AuditURL, "URL for audit log sending")

	return AuditURL
}

// profilingEnableParameter возвращает флаг включения профилирования.
func profilingEnableParameter() bool {
	profilingEnable := profilingEnable

	profilingEnableEnv, ok := os.LookupEnv("PROFILE")

	if ok {
		if val, err := strconv.ParseBool(profilingEnableEnv); err == nil {
			profilingEnable = val
		}
	}

	flag.BoolVar(&profilingEnable, "profile", profilingEnable, "Enable profiling")

	return profilingEnable
}

// profileServerAddressParameter возвращает адрес сервера профилирования.
func profileServerAddressParameter() string {
	profileServer := ""

	if profileServerEnv, ok := os.LookupEnv("PROFILE_ADDR"); ok {
		return profileServerEnv
	}

	flag.StringVar(&profileServer, "profile-addr", profileServer, "Address for pprof server")

	return profileServer
}

// profileDirParameter возвращает директорию для сохранения профилей.
func profileDirParameter() string {
	profileDir := ""

	if ProfileDirEnv, ok := os.LookupEnv("PROFILE_DIR"); ok {
		return ProfileDirEnv
	}

	flag.StringVar(&profileDir, "profile-dir", profileDir, "Address for pprof server")

	return profileDir
}

// cryptoKeyParameter возвращает путь до файла с крипто ключом из переменных окружения или флагов.
func cryptoKeyParameter() string {
	cryptoKey := ""

	if env, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		cryptoKey = env
	}
	flag.StringVar(&cryptoKey, "crypto-key", cryptoKey, "Crypto key")

	return cryptoKey
}

// configFileParameter возвращает путь до файла с конфигурационными параметрами из переменных окружения или флагов.
func configFileParameter() string {
	configFile := ""

	if env, ok := os.LookupEnv("CONFIG"); ok {
		configFile = env
	}

	flag.StringVar(&configFile, "c", configFile, "Config file")
	flag.StringVar(&configFile, "config", configFile, "Config file")

	return configFile
}

func trustedSubnetParameter() string {
	trustedSubnet := ""

	if env, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		trustedSubnet = env
	}
	flag.StringVar(&trustedSubnet, "t", trustedSubnet, "Trusted subnet in CIDR format (e.g. 192.168.1.0/24)")

	return trustedSubnet
}

func useGRPCParameter() bool {
	useGRPC := false

	if env, ok := os.LookupEnv("USE_GRPC"); ok {
		useGRPC, _ = strconv.ParseBool(env)
	}

	flag.BoolVar(&useGRPC, "grpc", useGRPC, "Use gRPC protocol")

	return useGRPC
}

func grpcAddressParameter() string {
	grpcAddr := ""

	if env, ok := os.LookupEnv("GRPC_ADDRESS"); ok {
		grpcAddr = env
	}

	flag.StringVar(&grpcAddr, "grpc-addr", grpcAddr, "gRPC server address")

	return grpcAddr
}
