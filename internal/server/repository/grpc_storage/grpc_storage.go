package grpcstorage

import (
	"context"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/proto"
	"github.com/Ko4etov/go-metrics/internal/server/repository/storage"
	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcMetricsStorage struct {
	proto.UnimplementedMetricsServer
	storage       *storage.MetricsStorage
}

// NewGrpcMetricsStorage создает новый сервер метрик для gRPC.
func NewGrpcMetricsStorage(storage *storage.MetricsStorage) *GrpcMetricsStorage {
	return &GrpcMetricsStorage{
		storage:       storage,
	}
}

// UpdateMetrics обрабатывает запрос на обновление метрик.
func (s *GrpcMetricsStorage) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	// Конвертируем protobuf в модели
	metrics := make([]models.Metrics, 0, len(req.Metrics))
	
	for _, pm := range req.Metrics {
		m := models.Metrics{
			ID:    pm.Id,
			MType: convertTypeBack(pm.Type),
			Hash:  pm.Hash,
		}
		
		switch pm.Type {
		case proto.Metric_COUNTER:
			m.Delta = &pm.Delta
		case proto.Metric_GAUGE:
			m.Value = &pm.Value
		}
		
		metrics = append(metrics, m)
	}

	// Сохраняем в хранилище
	if err := s.storage.UpdateMetricsBatch(metrics); err != nil {
		logger.Logger.Errorf("Failed to save metrics: %v", err)
		return nil, status.Error(codes.Internal, "failed to save metrics")
	}

	return &proto.UpdateMetricsResponse{}, nil
}

// convertTypeBack конвертирует protobuf enum в строковый тип.
func convertTypeBack(t proto.Metric_MType) string {
	switch t {
	case proto.Metric_COUNTER:
		return "counter"
	default:
		return "gauge"
	}
}