package handler

import (
	"fmt"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/repository/storage"
)

func GetMetrics(w http.ResponseWriter, r *http.Request) {
	storage := storage.New()
	metrics := storage.Metrics()

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, "<html><body>")
	fmt.Fprintln(w, "<h1>Metrics</h1>")

	fmt.Fprintln(w, "<ul>")
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value != nil {
				fmt.Fprintf(w, "<li>%s: %.5f</li>\n", metric.ID, *metric.Value)
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
