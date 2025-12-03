package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const address string = ":8080"
const reportInterval int = 2
const pollInterval int = 10
const rateLimit int = 1

type AgentParameters struct {
	Address string
	ReportInterval int
	PollInterval int
	HashKey string
	RateLimit int

}

func parseAgentParameters() *AgentParameters {
	godotenv.Load()
	
	addressParameter := addressParameter()
	reportIntervalParameter := reportIntervalParameter()
	pollIntervalParameter := pollIntervalParameter()
	hashKeyParameter := hashKeyParameter()
	rateLimitParameter := rateLimitParameter()

	flag.Parse()

	return &AgentParameters{
		Address: addressParameter,
		ReportInterval: reportIntervalParameter,
		PollInterval: pollIntervalParameter,
		HashKey: hashKeyParameter,
		RateLimit: rateLimitParameter,
	}
}


func rateLimitParameter() int {
	rateLimit := rateLimit

	if env, exist := os.LookupEnv("RATE_LIMIT"); exist {
		val, err := strconv.Atoi(env)

		if err != nil {
			os.Exit(1)
		}

		rateLimit = val
	}

	flag.IntVar(&rateLimit, "l", rateLimit, "Hash key")

	return rateLimit
}

func hashKeyParameter() string {
	hashKey := ""

	if env, exist := os.LookupEnv("KEY"); exist {
		hashKey = env
	}

	flag.StringVar(&hashKey, "k", hashKey, "Hash key")

	return hashKey
}

func addressParameter() string {
	address := address

	if addressEnv, exist := os.LookupEnv("ADDRESS"); exist {
		address = addressEnv
		return address
	}

	flag.StringVar(&address, "a", address, "Server address")

	return address
}

func reportIntervalParameter() int {
	reportInterval := reportInterval

	reportIntervalEnv, exist := os.LookupEnv("REPORT_INTERVAL")

	if !exist {
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

	pollIntervalEnv, exist := os.LookupEnv("POLL_INTERVAL")

	if !exist {
		flag.IntVar(&pollInterval, "p", pollInterval, "Poll interval in seconds")
		return pollInterval
	}

	result, err := strconv.ParseInt(pollIntervalEnv, 0, 64)

	if err != nil {
		os.Exit(1)
	}

	return int(result)
}