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
var hashKey string

type ServerParameters struct {
	Address string
	StoreMetricsInterval int
	FileStorageMetricsPath string
	RestoreMetrics bool
	DBAddress string
	HashKey string
}

func parseServerParameters() *ServerParameters {
	godotenv.Load()
	
	addressParameter()
	storeMetricsIntervalParameter()
	fileStorageMetricsPathParameter()
	restoreMetricsParameter()
	dbAddressParameter()
	hashKeyParameter()

	flag.Parse()

	logger.Logger.Infof("address=%v, storeMetricsInterval=%v, fileStorageMetricsPath=%v, restoreMetrics=%v, dbAddress=%v", address, storeMetricsInterval, fileStorageMetricsPath, restoreMetrics, dbAddress)

	return &ServerParameters{
		Address: address,
		StoreMetricsInterval: storeMetricsInterval,
		FileStorageMetricsPath: fileStorageMetricsPath,
		RestoreMetrics: restoreMetrics,
		DBAddress: dbAddress,
		HashKey: hashKey,
	}
}

func hashKeyParameter() {
	if env := os.Getenv("KEY"); env != "" {
		hashKey = env
	}

	flag.StringVar(&hashKey, "k", hashKey, "Hash key")
}

func dbAddressParameter() {
	if env := os.Getenv("DATABASE_DSN"); env != "" {
		dbAddress = env
	}

	flag.StringVar(&dbAddress, "d", dbAddress, "DB address")
}

func addressParameter() {
	if env := os.Getenv("ADDRESS"); env != "" {
		address = env
	}
	flag.StringVar(&address, "a", address, "Server address")
}

func storeMetricsIntervalParameter() {
	if env := os.Getenv("STORE_INTERVAL"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			storeMetricsInterval = val
		}
	}
	flag.IntVar(&storeMetricsInterval, "i", storeMetricsInterval, "store metrics interval in seconds")
}

func fileStorageMetricsPathParameter() {
	if fileStorageMetricsPathEnv := os.Getenv("FILE_STORAGE_PATH"); fileStorageMetricsPathEnv != "" {
		fileStorageMetricsPath = fileStorageMetricsPathEnv
		return
	}

	flag.StringVar(&fileStorageMetricsPath, "f", fileStorageMetricsPath, "file storage path")
}

func restoreMetricsParameter() {
	if env := os.Getenv("RESTORE"); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			restoreMetrics = val
		}
	}
	flag.BoolVar(&restoreMetrics, "r", restoreMetrics, "restore metrics")
}