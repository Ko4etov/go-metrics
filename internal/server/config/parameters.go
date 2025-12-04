package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const address string = ":8080"
const storeMetricsInterval int = 300
const fileStorageMetricsPath string = "metrics.json"
const restoreMetrics bool = true

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
	
	addressParameter := addressParameter()
	storeMetricsIntervalParameter := storeMetricsIntervalParameter()
	fileStorageMetricsPathParameter := fileStorageMetricsPathParameter()
	restoreMetricsParameter := restoreMetricsParameter()
	dbAddressParameter := dbAddressParameter()
	hashKeyParameter := hashKeyParameter()

	flag.Parse()

	return &ServerParameters{
		Address: addressParameter,
		StoreMetricsInterval: storeMetricsIntervalParameter,
		FileStorageMetricsPath: fileStorageMetricsPathParameter,
		RestoreMetrics: restoreMetricsParameter,
		DBAddress: dbAddressParameter,
		HashKey: hashKeyParameter,
	}
}

func hashKeyParameter() string {
	env, exist := os.LookupEnv("KEY")

	if !exist {
		// Значения по умолчанию нет, значит выходим
		os.Exit(2)
	}

	flag.StringVar(&env, "k", env, "Hash key")

	return env
}

func dbAddressParameter() string {
	dbAddress := ""

	if env, exist := os.LookupEnv("DATABASE_DSN"); exist {
		dbAddress = env
	}

	flag.StringVar(&dbAddress, "d", dbAddress, "DB address")

	return dbAddress
}

func addressParameter() string {
	address := address

	if env, exist := os.LookupEnv("ADDRESS"); exist {
		address = env
	}
	flag.StringVar(&address, "a", address, "Server address")

	return address
}

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

func fileStorageMetricsPathParameter() string {
	fileStorageMetricsPath := fileStorageMetricsPath

	if fileStorageMetricsPathEnv, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		fileStorageMetricsPath = fileStorageMetricsPathEnv
		return fileStorageMetricsPath
	}

	flag.StringVar(&fileStorageMetricsPath, "f", fileStorageMetricsPath, "file storage path")

	return fileStorageMetricsPath
}

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