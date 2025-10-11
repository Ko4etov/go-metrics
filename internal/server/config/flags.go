package config

import "flag"

var serverAddress string = ":8080"

func parseServerFlags() {
	flag.StringVar(&serverAddress, "a", serverAddress, "Server address")

	flag.Parse()
}