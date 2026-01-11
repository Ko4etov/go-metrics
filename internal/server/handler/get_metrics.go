package handler

import (
	"fmt"
	"net/http"
	"sort"
	"text/template"
)

// ViewData содержит данные для отображения страницы с метриками.
type ViewData struct {
	Title   string             // заголовок страницы
	Metrics []MetricsRecource  // список метрик
}

// MetricsRecource представляет метрику для отображения в HTML.
type MetricsRecource struct {
	Name  string // имя метрики
	Value string // значение метрики
}

// GetMetrics возвращает HTML-страницу со списком всех метрик.
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.storage.Metrics()

	w.Header().Set("Content-Type", "text/html")

	var MetricsSlice []MetricsRecource

	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value != nil {
				MetricsSlice = append(MetricsSlice, MetricsRecource{
					Name:  metric.ID,
					Value: fmt.Sprintf("%.2f", *metric.Value),
				})
			}
		case "counter":
			if metric.Delta != nil {
				MetricsSlice = append(MetricsSlice, MetricsRecource{
					Name:  metric.ID,
					Value: fmt.Sprintf("%d", *metric.Delta),
				})
			}
		}
	}

	sort.Slice(MetricsSlice, func(i, j int) bool {
		return MetricsSlice[i].Name < MetricsSlice[j].Name
	})

	tmpl, _ := template.ParseFiles("internal/server/templates/metrics.html")
	tmpl.Execute(w, MetricsSlice)
}