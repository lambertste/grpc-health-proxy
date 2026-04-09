package health

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type mockHealthServer struct {
	healthpb.UnimplementedHealthServer
	status healthpb.HealthCheckResponse_ServingStatus
}

func (m *mockHealthServer) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: m.status}, nil
}

func startMockServer(t *testing.T, status healthpb.HealthCheckResponse_ServingStatus) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	healthpb.RegisterHealthServer(grpcServer, &mockHealthServer{status: status})

	go func() {
		grpcServer.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		grpcServer.Stop()
	}
}

func TestNewChecker(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		timeout time.Duration
		wantErr bool
	}{
		{"valid", "localhost:8080", 5 * time.Second, false},
		{"empty target", "", 5 * time.Second, true},
		{"zero timeout", "localhost:8080", 0, true},
		{"negative timeout", "localhost:8080", -1 * time.Second, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewChecker(tt.target, tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewChecker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChecker_Check(t *testing.T) {
	addr, cleanup := startMockServer(t, healthpb.HealthCheckResponse_SERVING)
	defer cleanup()

	checker, err := NewChecker(addr, 5*time.Second)
	if err != nil {
		t.Fatalf("NewChecker() failed: %v", err)
	}

	if err := checker.Connect(); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer checker.Close()

	ctx := context.Background()
	status, err := checker.Check(ctx, "")
	if err != nil {
		t.Errorf("Check() error = %v", err)
	}
	if status != StatusServing {
		t.Errorf("Check() status = %v, want %v", status, StatusServing)
	}
}
