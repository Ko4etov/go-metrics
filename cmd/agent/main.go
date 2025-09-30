package main

import (
	"flag"
	"log"
	"time"

	"github.com/Ko4etov/go-metrics/internal/agent"
)

func main() {
    // Парсинг флагов командной строки
    pollInterval := flag.Int("p", 2, "Poll interval in seconds")
    reportInterval := flag.Int("r", 10, "Report interval in seconds")
    serverAddress := flag.String("a", "localhost:8080", "Server address")
    flag.Parse()

    // Создание агента
    ag := agent.NewAgent(
        time.Duration(*pollInterval)*time.Second,
        time.Duration(*reportInterval)*time.Second,
        *serverAddress,
    )

    log.Printf("Starting agent with pollInterval: %ds, reportInterval: %ds, server: %s",
        *pollInterval, *reportInterval, *serverAddress)

    // Запуск агента
    ag.Run()
}