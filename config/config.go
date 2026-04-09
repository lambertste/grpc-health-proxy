package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	// HTTP server settings
	HTTPPort int
	HTTPHost string

	// gRPC target settings
	GRPCTarget string
	GRPCTimeout time.Duration

	// Health check settings
	HealthCheckPath string
	ServiceName     string
}

// Load reads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:        getEnvAsInt("HTTP_PORT", 8080),
		HTTPHost:        getEnv("HTTP_HOST", "0.0.0.0"),
		GRPCTarget:      getEnv("GRPC_TARGET", "localhost:50051"),
		GRPCTimeout:     getEnvAsDuration("GRPC_TIMEOUT", 5*time.Second),
		HealthCheckPath: getEnv("HEALTH_CHECK_PATH", "/health"),
		ServiceName:     getEnv("SERVICE_NAME", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.HTTPPort)
	}
	if c.GRPCTarget == "" {
		return fmt.Errorf("GRPC_TARGET cannot be empty")
	}
	if c.HealthCheckPath == "" {
		return fmt.Errorf("HEALTH_CHECK_PATH cannot be empty")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
