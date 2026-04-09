package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnv := map[string]string{
		"HTTP_PORT":         os.Getenv("HTTP_PORT"),
		"HTTP_HOST":         os.Getenv("HTTP_HOST"),
		"GRPC_TARGET":       os.Getenv("GRPC_TARGET"),
		"GRPC_TIMEOUT":      os.Getenv("GRPC_TIMEOUT"),
		"HEALTH_CHECK_PATH": os.Getenv("HEALTH_CHECK_PATH"),
		"SERVICE_NAME":      os.Getenv("SERVICE_NAME"),
	}
	defer func() {
		for k, v := range originalEnv {
			os.Setenv(k, v)
		}
	}()

	// Clear env vars
	for k := range originalEnv {
		os.Unsetenv(k)
	}

	t.Run("default configuration", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cfg.HTTPPort != 8080 {
			t.Errorf("expected HTTPPort 8080, got %d", cfg.HTTPPort)
		}
		if cfg.GRPCTarget != "localhost:50051" {
			t.Errorf("expected GRPCTarget localhost:50051, got %s", cfg.GRPCTarget)
		}
	})

	t.Run("custom configuration", func(t *testing.T) {
		os.Setenv("HTTP_PORT", "9090")
		os.Setenv("GRPC_TARGET", "myservice:50051")
		os.Setenv("GRPC_TIMEOUT", "10s")
		os.Setenv("SERVICE_NAME", "my-service")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cfg.HTTPPort != 9090 {
			t.Errorf("expected HTTPPort 9090, got %d", cfg.HTTPPort)
		}
		if cfg.GRPCTarget != "myservice:50051" {
			t.Errorf("expected GRPCTarget myservice:50051, got %s", cfg.GRPCTarget)
		}
		if cfg.GRPCTimeout != 10*time.Second {
			t.Errorf("expected GRPCTimeout 10s, got %v", cfg.GRPCTimeout)
		}
		if cfg.ServiceName != "my-service" {
			t.Errorf("expected ServiceName my-service, got %s", cfg.ServiceName)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("invalid port", func(t *testing.T) {
		cfg := &Config{HTTPPort: 70000, GRPCTarget: "localhost:50051", HealthCheckPath: "/health"}
		if err := cfg.Validate(); err == nil {
			t.Error("expected validation error for invalid port")
		}
	})

	t.Run("empty grpc target", func(t *testing.T) {
		cfg := &Config{HTTPPort: 8080, GRPCTarget: "", HealthCheckPath: "/health"}
		if err := cfg.Validate(); err == nil {
			t.Error("expected validation error for empty GRPC target")
		}
	})
}
