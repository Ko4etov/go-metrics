package main

import (
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metric_type}/{metric_name}/{metric_value}", handler.UpdateMetricHandler)
	mux.HandleFunc("/metrics/", handler.GetMetricsHandler)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
        panic(err)
    }
}
