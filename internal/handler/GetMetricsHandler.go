package handler

import (
	"fmt"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/storage"
)

func GetMetricsHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    storage := storage.GetInstance()
    metrics := storage.GetAllMetrics()

    w.Header().Set("Content-Type", "text/html")
    fmt.Fprintln(w, "<html><body>")
    fmt.Fprintln(w, "<h1>Metrics</h1>")
    
    fmt.Fprintln(w, "<ul>")
    for name, value := range metrics {
        fmt.Fprintf(w, "<li>%s: %f</li>\n", name, value)
    }
    fmt.Fprintln(w, "</ul>")
    
    fmt.Fprintln(w, "</body></html>")
}