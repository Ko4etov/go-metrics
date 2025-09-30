package handler

import (
	"fmt"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/storage"
)

func GetMetricsHandler(w http.ResponseWriter, r *http.Request) {

    storage := storage.GetInstance()
    metrics := storage.GetAllMetrics()

    w.Header().Set("Content-Type", "text/html")
    fmt.Fprintln(w, "<html><body>")
    fmt.Fprintln(w, "<h1>Metrics</h1>")
    
    fmt.Fprintln(w, "<ul>")
    for _ , metric := range metrics {
        switch metric.MType {
        case "gauge":
            if metric.Value != nil {
                fmt.Fprintf(w, "<li>%s: %.2f</li>\n", metric.ID, *metric.Value)
            }
        case "counter":
            if metric.Delta != nil {
                fmt.Fprintf(w, "<li>%s: %d</li>\n", metric.ID, *metric.Delta)
            }
        }
    }
    fmt.Fprintln(w, "</ul>")
    
    fmt.Fprintln(w, "</body></html>")
}