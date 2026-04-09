package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/yourorg/grpc-health-proxy/metrics"
)

func TestNew_RegistersMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.New(reg)

	if m.HealthCheckTotal == nil {
		t.Fatal("expected HealthCheckTotal to be non-nil")
	}
	if m.HealthCheckLatency == nil {
		t.Fatal("expected HealthCheckLatency to be non-nil")
	}
}

func TestNew_PanicsOnDoubleRegister(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics.New(reg)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on double registration")
		}
	}()
	metrics.New(reg)
}

func TestHandler_ExposesMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := metrics.New(reg)

	m.HealthCheckTotal.WithLabelValues("my-service", "ok").Inc()
	m.HealthCheckLatency.WithLabelValues("my-service").Observe(0.042)

	h := metrics.Handler(reg)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "grpc_health_proxy_health_checks_total") {
		t.Error("expected metrics output to contain health_checks_total")
	}
	if !strings.Contains(string(body), "grpc_health_proxy_health_check_duration_seconds") {
		t.Error("expected metrics output to contain health_check_duration_seconds")
	}
}
