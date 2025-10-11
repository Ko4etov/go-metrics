package config

import "flag"

var agentAddress string = ":8080"
var pollInterval int = 2
var reportInterval int = 10

func parseAgentFlags() {
	flag.IntVar(&pollInterval, "p", pollInterval, "Poll interval in seconds")
	flag.IntVar(&reportInterval, "r", reportInterval, "Report interval in seconds")
	flag.StringVar(&agentAddress, "a", agentAddress, "Agent address")

	flag.Parse()
}