package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
	"github.com/joho/godotenv"
)

var address string = ":8080"
var storeMetricsInterval int = 300
var fileStorageMetricsPath string = "metrics.json"
var restoreMetrics bool = true
var dbAddress string

type ServerParameters struct {
	Address string
	StoreMetricsInterval int
	FileStorageMetricsPath string
	RestoreMetrics bool
	DBAddress string
}

func parseServerParameters() *ServerParameters {
	err := godotenv.Load()
	if err != nil {
		logger.Logger.Infof("Error loading .env file: %v", err)
	}
	
	addressParameter()
	storeMetricsIntervalParameter()
	fileStorageMetricsPathParameter()
	restoreMetricsParameter()
	dbAddressParameter()

	flag.Parse()

	logger.Logger.Infof("address=%v, storeMetricsInterval=%v, fileStorageMetricsPath=%v, restoreMetrics=%v, dbAddress=%v", address, storeMetricsInterval, fileStorageMetricsPath, restoreMetrics, dbAddress)

	return &ServerParameters{
		Address: address,
		StoreMetricsInterval: storeMetricsInterval,
		FileStorageMetricsPath: fileStorageMetricsPath,
		RestoreMetrics: restoreMetrics,
		DBAddress: dbAddress,
	}
}

func dbAddressParameter() {
	if dbAddressEnv := os.Getenv("DATABASE_DSN"); dbAddressEnv != "" {
		dbAddress = dbAddressEnv
		return
	}

	flag.StringVar(&dbAddress, "d", dbAddress, "DB address")
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

	if result, err := strconv.Atoi(storeMetricsIntervalEnv); err == nil {
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