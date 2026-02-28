package grpcsender

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Ko4etov/go-metrics/internal/models"
	"github.com/Ko4etov/go-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GRPCSender struct {
	client     proto.MetricsClient
	conn       *grpc.ClientConn
	serverAddr string
	hashKey    string
	localIP    string
}

func New(serverAddr, hashKey string) (*GRPCSender, error) {
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := proto.NewMetricsClient(conn)
	
	localIP, _ := getLocalIP()

	return &GRPCSender{
		client:     client,
		conn:       conn,
		serverAddr: serverAddr,
		hashKey:    hashKey,
		localIP:    localIP,
	}, nil
}

func (s *GRPCSender) SendMetrics(metrics []models.Metrics) {
	if len(metrics) == 0 {
		return
	}

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

	ctx := metadata.NewOutgoingContext(context.Background(),
		metadata.Pairs("x-real-ip", s.localIP))

	_, err := s.client.UpdateMetrics(ctx, req)
	if err != nil {
		log.Printf("Failed to send metrics via gRPC: %v", err)
	}
}

func (s *GRPCSender) Stop() {
	if s.conn != nil {
		s.conn.Close()
	}
}

func convertType(mType string) proto.Metric_MType {
	switch mType {
	case "counter":
		return proto.Metric_COUNTER
	default:
		return proto.Metric_GAUGE
	}
}

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