package grpcsender

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCSender отправляет метрики по gRPC.
type GRPCSender struct {
	client     proto.MetricsClient
	conn       *grpc.ClientConn
	serverAddr string
	hashKey    string
	localIP    string
}

// New создает новый gRPC отправитель.
func New(serverAddr, hashKey string) (*GRPCSender, error) {
	conn, err := grpc.Dial(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := proto.NewMetricsClient(conn)
	
	// Получаем локальный IP
	localIP, _ := getLocalIP()

	return &GRPCSender{
		client:     client,
		conn:       conn,
		serverAddr: serverAddr,
		hashKey:    hashKey,
		localIP:    localIP,
	}, nil
}

// SendMetrics отправляет метрики на сервер.
func (s *GRPCSender) SendMetrics(metrics []models.Metrics) {
	if len(metrics) == 0 {
		return
	}

	// Конвертируем модели в protobuf
	protoMetrics := make([]*proto.Metric, 0, len(metrics))
	for _, m := range metrics {
		pm := &proto.Metric{
			Id:   m.ID,
			Type: convertType(m.MType),
			Hash: m.Hash,
		}
		
		if m.Delta != nil {
			pm.Delta = *m.Delta
		}
		if m.Value != nil {
			pm.Value = *m.Value
		}
		
		protoMetrics = append(protoMetrics, pm)
	}

	req := &proto.UpdateMetricsRequest{
		Metrics: protoMetrics,
	}

	// Добавляем IP в метаданные
	ctx := metadata.NewOutgoingContext(context.Background(),
		metadata.Pairs("x-real-ip", s.localIP))

	// Отправляем запрос
	_, err := s.client.UpdateMetrics(ctx, req)
	if err != nil {
		log.Printf("Failed to send metrics via gRPC: %v", err)
	}
}

// Stop закрывает соединение.
func (s *GRPCSender) Stop() {
	if s.conn != nil {
		s.conn.Close()
	}
}

// convertType конвертирует тип метрики в protobuf enum.
func convertType(mType string) proto.Metric_MType {
	switch mType {
	case "counter":
		return proto.Metric_COUNTER
	default:
		return proto.Metric_GAUGE
	}
}

// getLocalIP получает локальный IP адрес.
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IP address found")
}