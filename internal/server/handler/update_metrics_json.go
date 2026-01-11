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
) ([]string, int, error) {

	if req.Header.Get("Content-Type") != "application/json" {
		return nil, http.StatusBadRequest, fmt.Errorf("Content-Type must be application/json")
	}

	metrics := make([]models.Metrics, 0, 10)
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err)
	}

	if len(metrics) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empty metrics batch")
	}

	for i := range metrics {
		if err := h.validateMetric(&metrics[i]); err != nil {
			return nil, http.StatusBadRequest, fmt.Errorf("invalid metric: %w", err)
		}
	}

	var metricNames []string
	if auditSvc != nil {
		metricNames = make([]string, 0, len(metrics))

		for i := range metrics {
			metricNames = append(metricNames, metrics[i].ID)
		}
	}

	if err := h.storage.UpdateMetricsBatch(metrics); err != nil {
		return metricNames, http.StatusInternalServerError,
			fmt.Errorf("failed to update metrics: %w", err)
	}

	return metricNames, http.StatusOK, nil
}

func (h *Handler) UpdateMetricsBatch(res http.ResponseWriter, req *http.Request) {
	_, statusCode, err := h.processMetricsBatchInternal(res, req, nil)

	if err != nil {
		http.Error(res, err.Error(), statusCode)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateMetricsBatchWithAudit(auditSvc *audit.AuditService) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metricNames, statusCode, err := h.processMetricsBatchInternal(res, req, auditSvc)

		if err != nil {
			http.Error(res, err.Error(), statusCode)
			return
		}

		if auditSvc != nil && statusCode == http.StatusOK && len(metricNames) > 0 {
			go h.sendAuditEvent(req, metricNames, auditSvc)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) sendAuditEvent(req *http.Request, metricNames []string, auditSvc *audit.AuditService) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: getIPAddress(req),
	}

	if err := auditSvc.Notify(ctx, event); err != nil {
		fmt.Printf("[audit] Failed to send audit event: %v\n", err)
	}
}
