package healthz_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/grpc-health-proxy/healthz"
)

func TestLive_AlwaysReturns200(t *testing.T) {
	h := healthz.New()

	for _, ready := range []bool{false, true} {
		h.SetReady(ready)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/healthz/live", nil)
		h.Live(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("ready=%v: expected 200, got %d", ready, rec.Code)
		}
	}
}

func TestReady_Returns503WhenNotReady(t *testing.T) {
	h := healthz.New()
	// default: not ready

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz/ready", nil)
	h.Ready(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}

	var s map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if s["status"] != "not_ready" {
		t.Errorf("expected status=not_ready, got %v", s["status"])
	}
}

func TestReady_Returns200WhenReady(t *testing.T) {
	h := healthz.New()
	h.SetReady(true)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz/ready", nil)
	h.Ready(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var s map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if s["status"] != "ready" {
		t.Errorf("expected status=ready, got %v", s["status"])
	}
}

func TestRegister_MountsEndpoints(t *testing.T) {
	h := healthz.New()
	h.SetReady(true)
	mux := http.NewServeMux()
	h.Register(mux)

	for _, path := range []string{"/healthz/live", "/healthz/ready"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", path, rec.Code)
		}
	}
}

func TestLive_ResponseContainsUptimeAndTimestamp(t *testing.T) {
	h := healthz.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz/live", nil)
	h.Live(rec, req)

	var s map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if _, ok := s["uptime"]; !ok {
		t.Error("expected uptime field in response")
	}
	if _, ok := s["timestamp"]; !ok {
		t.Error("expected timestamp field in response")
	}
}
