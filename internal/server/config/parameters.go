package config

import (
	"flag"
	"os"
	"strconv"
)

var address string = ":8080"
var storeMetricsInterval int = 300
var fileStorageMetricsPath string = "metrics.json"
var restoreMetrics bool

func parseServerParameters() {
	addressParameter()
	storeMetricsIntervalParameter()
	fileStorageMetricsPathParameter()
	restoreMetricsParameter()

	flag.Parse()
}

func addressParameter() {
	if addressEnv := os.Getenv("ADDRESS"); addressEnv != "" {
		address = addressEnv
		return
	}

	flag.StringVar(&address, "a", address, "Server address")
}


func storeMetricsIntervalParameter() {
	storeMetricsIntervalEnv := os.Getenv("STORE_INTERVAL")

	if storeMetricsIntervalEnv == "" {
		flag.IntVar(&storeMetricsInterval, "i", storeMetricsInterval, "store metrics interval in seconds")
		return
	}

	if result, err := strconv.ParseInt(storeMetricsIntervalEnv, 0, 64); err == nil {
		storeMetricsInterval = int(result)
	}
}

func fileStorageMetricsPathParameter() {
	if fileStorageMetricsPathEnv := os.Getenv("FILE_STORAGE_PATH"); fileStorageMetricsPathEnv != "" {
		fileStorageMetricsPath = fileStorageMetricsPathEnv
		return
	}

	flag.StringVar(&fileStorageMetricsPath, "f", fileStorageMetricsPath, "file storage path")
}

func restoreMetricsParameter() {
	restoreMetricsEnv := os.Getenv("RESTORE")

	if restoreMetricsEnv == "" {
		flag.BoolVar(&restoreMetrics, "r", restoreMetrics, "restore metrics")
		return
	}

	if result, err := strconv.ParseBool(restoreMetricsEnv); err == nil {
		restoreMetrics = bool(result)
	}
}