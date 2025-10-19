package main

import (
	"github.com/Ko4etov/go-metrics/internal/server"
	"github.com/Ko4etov/go-metrics/internal/server/config"
)

func main() {
	config := config.New()

	server.New(config.ServerAddress).Run()
}
