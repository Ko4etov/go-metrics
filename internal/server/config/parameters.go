// Package config содержит параметры конфигурации сервера сбора метрик.
package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

const (
	address                 = ":8080"           // Адрес сервера по умолчанию
	storeMetricsInterval    = 300               // Интервал сохранения метрик по умолчанию
	fileStorageMetricsPath  = "metrics.json"    // Путь к файлу метрик по умолчанию
	restoreMetrics          = true              // Восстанавливать метрики по умолчанию
	profilingEnable         = false             // Профилирование отключено по умолчанию
)

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
}

// parseServerParameters парсит параметры сервера из переменных окружения и флагов.
func parseServerParameters() *ServerParameters {
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info(".env file not loaded: %v", err)
	}

	addressParameter := addressParameter()
	storeMetricsIntervalParameter := storeMetricsIntervalParameter()
	fileStorageMetricsPathParameter := fileStorageMetricsPathParameter()
	restoreMetricsParameter := restoreMetricsParameter()
	dbAddressParameter := dbAddressParameter()
	hashKeyParameter := hashKeyParameter()
	auditFileParameter := auditFileParameter()
	AuditURLParameter := auditURLParameter()
	profilingEnableParameter := profilingEnableParameter()
	profileServerParameter := profileServerAddressParameter()
	profileDirParameter := profileDirParameter()

	flag.Parse()

	return &ServerParameters{
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
	}
}

// hashKeyParameter возвращает ключ для хеширования из переменных окружения или флагов.
func hashKeyParameter() string {
	env, exist := os.LookupEnv("KEY")

	if !exist {
		os.Exit(2)
	}

	flag.StringVar(&env, "k", env, "Hash key")

	return env
}

// dbAddressParameter возвращает адрес базы данных из переменных окружения или флагов.
func dbAddressParameter() string {
	dbAddress := ""

	if env, exist := os.LookupEnv("DATABASE_DSN"); exist {
		dbAddress = env
	}

	flag.StringVar(&dbAddress, "d", dbAddress, "DB address")

	return dbAddress
}

// addressParameter возвращает адрес сервера из переменных окружения или флагов.
func addressParameter() string {
	address := address

	if env, exist := os.LookupEnv("ADDRESS"); exist {
		address = env
	}
	flag.StringVar(&address, "a", address, "Server address")

	return address
}

// storeMetricsIntervalParameter возвращает интервал сохранения метрик.
func storeMetricsIntervalParameter() int {
	storeMetricsInterval := storeMetricsInterval

	if storeMetricsIntervalEnv, exist := os.LookupEnv("STORE_INTERVAL"); exist {
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

	if fileStorageMetricsPathEnv, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		fileStorageMetricsPath = fileStorageMetricsPathEnv
		return fileStorageMetricsPath
	}

	flag.StringVar(&fileStorageMetricsPath, "f", fileStorageMetricsPath, "file storage path")

	return fileStorageMetricsPath
}

// restoreMetricsParameter возвращает флаг восстановления метрик.
func restoreMetricsParameter() bool {
	restoreMetrics := restoreMetrics

	restoreMetricsEnv, exist := os.LookupEnv("RESTORE")

	if exist {
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

	if auditFileEnv, exist := os.LookupEnv("AUDIT_FILE"); exist {
		return auditFileEnv
	}

	flag.StringVar(&auditFile, "audit-file", auditFile, "Path to audit log file")

	return auditFile
}

// auditURLParameter возвращает URL для отправки аудита.
func auditURLParameter() string {
	AuditURL := ""

	if AuditURLEnv, exist := os.LookupEnv("AUDIT_URL"); exist {
		return AuditURLEnv
	}

	flag.StringVar(&AuditURL, "audit-url", AuditURL, "URL for audit log sending")

	return AuditURL
}

// profilingEnableParameter возвращает флаг включения профилирования.
func profilingEnableParameter() bool {
	profilingEnable := profilingEnable

	profilingEnableEnv, exist := os.LookupEnv("PROFILE")

	if exist {
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

	if profileServerEnv, exist := os.LookupEnv("PROFILE_ADDR"); exist {
		return profileServerEnv
	}

	flag.StringVar(&profileServer, "profile-addr", profileServer, "Address for pprof server")

	return profileServer
}

// profileDirParameter возвращает директорию для сохранения профилей.
func profileDirParameter() string {
	profileDir := ""

	if ProfileDirEnv, exist := os.LookupEnv("PROFILE_DIR"); exist {
		return ProfileDirEnv
	}

	flag.StringVar(&profileDir, "profile-dir", profileDir, "Address for pprof server")

	return profileDir
}