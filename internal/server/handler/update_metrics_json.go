package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/server/service/audit"
)

func getIPAddress(req *http.Request) string {
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return ip
}

func (h *Handler) processMetricsBatchInternal(
	res http.ResponseWriter, 
	req *http.Request,
	auditSvc *audit.AuditService,
) ([]string, int, error) { // возвращаем: имена метрик, статус код, ошибку
	
	// Валидация Content-Type
	if req.Header.Get("Content-Type") != "application/json" {
		return nil, http.StatusBadRequest, fmt.Errorf("Content-Type must be application/json")
	}

	// Декодирование JSON
	var metrics []models.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("Invalid JSON: %w", err)
	}

	// Проверка на пустой batch
	if len(metrics) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("Empty metrics batch")
	}

	// Валидация и сбор имен метрик
	var validMetrics []models.Metrics
	var metricNames []string
	for _, metric := range metrics {
		if err := h.validateMetric(&metric); err != nil {
			return nil, http.StatusBadRequest, fmt.Errorf("Invalid metric: %w", err)
		}
		validMetrics = append(validMetrics, metric)
		
		// Собираем имена только если нужен аудит
		if auditSvc != nil {
			metricNames = append(metricNames, metric.ID)
		}
	}

	// Сохранение метрик
	if err := h.storage.UpdateMetricsBatch(validMetrics); err != nil {
		// Возвращаем имена метрик даже при ошибке (если они были собраны)
		return metricNames, http.StatusInternalServerError, 
			fmt.Errorf("Failed to update metrics: %w", err)
	}

	return metricNames, http.StatusOK, nil
}

func (h *Handler) UpdateMetricsBatch(res http.ResponseWriter, req *http.Request) {
	// Вызываем общую функцию БЕЗ аудита
	_, statusCode, err := h.processMetricsBatchInternal(res, req, nil)
	
	if err != nil {
		http.Error(res, err.Error(), statusCode)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
}

// UpdateMetricsBatchWithAudit - НОВЫЙ МЕТОД с аудитом
func (h *Handler) UpdateMetricsBatchWithAudit(auditSvc *audit.AuditService) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Вызываем общую функцию С аудитом
		metricNames, statusCode, err := h.processMetricsBatchInternal(res, req, auditSvc)
		
		if err != nil {
			http.Error(res, err.Error(), statusCode)
			return
		}

		// ТОЛЬКО ПОСЛЕ УСПЕШНОГО СОХРАНЕНИЯ делаем аудит
		if auditSvc != nil && statusCode == http.StatusOK && len(metricNames) > 0 {
			go h.sendAuditEvent(req, metricNames, auditSvc)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
	}
}

// sendAuditEvent - отправка события аудита
func (h *Handler) sendAuditEvent(req *http.Request, metricNames []string, auditSvc *audit.AuditService) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: getIPAddress(req),
	}

	if err := auditSvc.Notify(ctx, event); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		fmt.Printf("[AUDIT] Failed to send audit event: %v\n", err)
	}
}