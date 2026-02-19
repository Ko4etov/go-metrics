package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
	"github.com/joho/godotenv"
)

const (
	address        string = ":8080" // адрес сервера по умолчанию
	reportInterval int    = 2       // интервал отправки метрик по умолчанию (в секундах)
	pollInterval   int    = 10      // интервал опроса метрик по умолчанию (в секундах)
	rateLimit      int    = 1       // лимит запросов по умолчанию
)

// AgentParameters содержит конфигурационные параметры для агента.
type AgentParameters struct {
	Address        string
	ReportInterval int
	PollInterval   int
	HashKey        string
	RateLimit      int
	CryptoKey      string
}

// parseAgentParameters парсит параметры агента.
func parseAgentParameters() *AgentParameters {
	godotenv.Load()

	addressParameter := addressParameter()
	reportIntervalParameter := reportIntervalParameter()
	pollIntervalParameter := pollIntervalParameter()
	hashKeyParameter := hashKeyParameter()
	rateLimitParameter := rateLimitParameter()
	cryptoKeyParameter := cryptoKeyParameter()

	flag.Parse()

	return &AgentParameters{
		Address:        addressParameter,
		ReportInterval: reportIntervalParameter,
		PollInterval:   pollIntervalParameter,
		HashKey:        hashKeyParameter,
		RateLimit:      rateLimitParameter,
		CryptoKey:      cryptoKeyParameter,
	}
}

func rateLimitParameter() int {
	rateLimit := rateLimit

	if env, ok := os.LookupEnv("RATE_LIMIT"); ok {
		val, err := strconv.Atoi(env)

		if err != nil {
			logger.Logger.Warn("can,t convert parameter Rate limit to integer")
		} else {
			rateLimit = val
		}
	}

	flag.IntVar(&rateLimit, "l", rateLimit, "Rate limit")

	return rateLimit
}

func hashKeyParameter() string {
	hashKey := ""

	if env, ok := os.LookupEnv("KEY"); ok {
		hashKey = env
	}

	flag.StringVar(&hashKey, "k", hashKey, "Hash key")

	return hashKey
}

func addressParameter() string {
	address := address

	if addressEnv, ok := os.LookupEnv("ADDRESS"); ok {
		address = addressEnv
		return address
	}

	flag.StringVar(&address, "a", address, "Server address")

	return address
}

func reportIntervalParameter() int {
	reportInterval := reportInterval

	reportIntervalEnv, ok := os.LookupEnv("REPORT_INTERVAL")

	if !ok {
		flag.IntVar(&reportInterval, "r", reportInterval, "Report interval in seconds")
		return reportInterval
	}

	if result, err := strconv.ParseInt(reportIntervalEnv, 0, 64); err == nil {
		reportInterval = int(result)
	}

	return reportInterval
}

func pollIntervalParameter() int {
	pollInterval := pollInterval

	pollIntervalEnv, ok := os.LookupEnv("POLL_INTERVAL")

	if !ok {
		flag.IntVar(&pollInterval, "p", pollInterval, "Poll interval in seconds")
		return pollInterval
	}

	result, err := strconv.ParseInt(pollIntervalEnv, 0, 64)

	if err != nil {
		logger.Logger.Warn("can,t convert parameter Poll interval to integer")
	} else {
		pollInterval = int(result)
	}

	return pollInterval
}

// cryptoKeyParameter возвращает путь до файла с крипто ключом из переменных окружения или флагов.
func cryptoKeyParameter() string {
	cryptoKey := ""

	if env, ok := os.LookupEnv("CRYPTO_KEY_AGENT"); ok {
		return env;
	}

	if env, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		cryptoKey = env
	}
	
	flag.StringVar(&cryptoKey, "crypto-key", cryptoKey, "Crypto key")

	return cryptoKey
}
