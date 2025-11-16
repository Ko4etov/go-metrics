package config

import (
	"flag"
	"os"
	"strconv"
)

var address string = ":8080"
var reportInterval int = 2
var pollInterval int = 10
var hashKey string

type AgentParameters struct {
	Address string
	ReportInterval int
	PollInterval int
	HashKey string
}

func parseAgentParameters() *AgentParameters {
	addressParameter()
	reportIntervalParameter()
	pollIntervalParameter()
	hashKeyParameter()

	flag.Parse()

	return &AgentParameters{
		Address: address,
		ReportInterval: reportInterval,
		PollInterval: pollInterval,
		HashKey: hashKey,
	}
}

func hashKeyParameter() {
	if env := os.Getenv("KEY"); env != "" {
		hashKey = env
	}

	flag.StringVar(&hashKey, "k", hashKey, "Hash key")
}

func addressParameter() {
	if addressEnv := os.Getenv("ADDRESS"); addressEnv != "" {
		address = addressEnv
		return
	}

	flag.StringVar(&address, "a", address, "Server address")
}

func reportIntervalParameter() {
	reportIntervalEnv := os.Getenv("REPORT_INTERVAL")

	if reportIntervalEnv == "" {
		flag.IntVar(&reportInterval, "r", reportInterval, "Report interval in seconds")
		return
	}

	if result, err := strconv.ParseInt(reportIntervalEnv, 0, 64); err == nil {
		reportInterval = int(result)
	}
}

func pollIntervalParameter() {
	pollIntervalEnv := os.Getenv("POLL_INTERVAL")

	if pollIntervalEnv == "" {
		flag.IntVar(&pollInterval, "p", pollInterval, "Poll interval in seconds")
		return
	}

	if result, err := strconv.ParseInt(pollIntervalEnv, 0, 64); err == nil {
		pollInterval = int(result)
	}
}