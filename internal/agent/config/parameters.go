package config

import (
	"flag"
	"os"
	"strconv"
)

var address string = ":8080"
var reportInterval int = 2
var pollInterval int = 10

func parseAgentParameters() {
	addressParameter()
	reportIntervalParameter()
	pollIntervalParameter()

	flag.Parse()
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