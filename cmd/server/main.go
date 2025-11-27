package main

import (
	"github.com/Ko4etov/go-metrics/internal/server"
	"github.com/Ko4etov/go-metrics/internal/server/config"
)

func main() {
	config, err := config.New()

	if err != nil {
		panic(err)
	}

	server.New(config).Run()
}
