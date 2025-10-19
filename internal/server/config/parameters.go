package config

import (
	"flag"
	"os"
)

var address string = ":8080"

func parseServerParameters() {
	addressParameter()

	flag.Parse()
}

func addressParameter() {
	if addressEnv := os.Getenv("ADDRESS"); addressEnv != "" {
		address = addressEnv
		return
	}

	flag.StringVar(&address, "a", address, "Server address")
}