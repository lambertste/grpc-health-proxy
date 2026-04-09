package health

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Status represents the health check status
type Status string

const (
	StatusServing    Status = "SERVING"
	StatusNotServing Status = "NOT_SERVING"
	StatusUnknown    Status = "UNKNOWN"
)

// Checker performs gRPC health checks
type Checker struct {
	target  string
	timeout time.Duration
	conn    *grpc.ClientConn
	client  healthpb.HealthClient
}

// NewChecker creates a new health checker
func NewChecker(target string, timeout time.Duration) (*Checker, error) {
	if target == "" {
		return nil, fmt.Errorf("target cannot be empty")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("timeout must be positive")
	}

	return &Checker{
		target:  target,
		timeout: timeout,
	}, nil
}

// Connect establishes a connection to the gRPC server
func (c *Checker) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.target, err)
	}

	c.conn = conn
	c.client = healthpb.NewHealthClient(conn)
	return nil
}

// Check performs a health check for the specified service
func (c *Checker) Check(ctx context.Context, service string) (Status, error) {
	if c.client == nil {
		return StatusUnknown, fmt.Errorf("not connected")
	}

	req := &healthpb.HealthCheckRequest{Service: service}
	resp, err := c.client.Check(ctx, req)
	if err != nil {
		return StatusUnknown, fmt.Errorf("health check failed: %w", err)
	}

	switch resp.Status {
	case healthpb.HealthCheckResponse_SERVING:
		return StatusServing, nil
	case healthpb.HealthCheckResponse_NOT_SERVING:
		return StatusNotServing, nil
	default:
		return StatusUnknown, nil
	}
}

// Close closes the gRPC connection
func (c *Checker) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
