package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"grpc-health-proxy/config"
	"grpc-health-proxy/health"
)

type mockChecker struct {
	status string
	err    error
}

func (m *mockChecker) Check(ctx context.Context, service string) (string, error) {
	return m.status, m.err
}

func (m *mockChecker) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		Timeout:  5 * time.Second,
	}
	checker := &mockChecker{status: "SERVING"}

	s := New(cfg, checker)
	if s == nil {
		t.Fatal("Expected server instance, got nil")
	}
	if s.config.HTTPPort != 8080 {
		t.Errorf("Expected port 8080, got %d", s.config.HTTPPort)
	}
}

func TestHandleHealth_Success(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		Timeout:  5 * time.Second,
	}
	checker := &mockChecker{status: "SERVING"}
	s := New(cfg, checker)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	if response["status"] != "SERVING" {
		t.Errorf("Expected status SERVING, got %s", response["status"])
	}
}

func TestHandleHealth_Failure(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		Timeout:  5 * time.Second,
	}
	checker := &mockChecker{status: "NOT_SERVING"}
	s := New(cfg, checker)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	if response["status"] != "NOT_SERVING" {
		t.Errorf("Expected status NOT_SERVING, got %s", response["status"])
	}
}

func TestHandleHealth_CheckerError(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		Timeout:  5 * time.Second,
	}
	checker := &mockChecker{err: errors.New("connection refused")}
	s := New(cfg, checker)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 on checker error, got %d", w.Code)
	}
}

func TestHandleReady(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		Timeout:  5 * time.Second,
	}
	checker := &mockChecker{status: "SERVING"}
	s := New(cfg, checker)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	s.handleReady(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
