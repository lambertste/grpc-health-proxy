package timeout_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/timeout"
)

func slowHandler(d time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			// context cancelled — do nothing so the timeout branch can reply
		}
	}
}

func TestNew_UsesDefaultTimeoutWhenZero(t *testing.T) {
	h := timeout.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Error("expected a deadline to be set")
		}
		if time.Until(deadline) > timeout.DefaultTimeout {
			t.Error("deadline exceeds DefaultTimeout")
		}
		w.WriteHeader(http.StatusOK)
	}), 0)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServeHTTP_PassesThroughFastHandler(t *testing.T) {
	h := timeout.New(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), 100*time.Millisecond)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestServeHTTP_Returns504OnTimeout(t *testing.T) {
	h := timeout.New(slowHandler(500*time.Millisecond), 50*time.Millisecond)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", rec.Code)
	}
	if body := rec.Body.String(); body == "" {
		t.Error("expected non-empty body for timeout response")
	}
}

func TestServeHTTP_BodyContainsDuration(t *testing.T) {
	d := 40 * time.Millisecond
	h := timeout.New(slowHandler(500*time.Millisecond), d)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	expected := d.String()
	if body := rec.Body.String(); !contains(body, expected) {
		t.Errorf("expected body to contain %q, got %q", expected, body)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
