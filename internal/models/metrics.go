// Package models содержит модели данных для системы метрик.
package models

const (
	// Counter - тип метрики "счетчик"
	Counter = "counter"
	// Gauge - тип метрики "измеритель"
	Gauge   = "gauge"
)

// Metrics представляет метрику системы.
// Delta и Value объявлены через указатели,
// чтобы отличать значение "0" от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // тип метрики (counter или gauge)
	Delta *int64   `json:"delta,omitempty"` // значение для счетчика (опционально)
	Value *float64 `json:"value,omitempty"` // значение для измерителя (опционально)
	Hash  string   `json:"hash,omitempty"`  // хеш для проверки целостности (опционально)
}