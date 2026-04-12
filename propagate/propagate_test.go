package propagate_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpc-health-proxy/propagate"
)

func captureCtxHandler(captured *http.Header) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = propagate.Headers(r.Context())
		w.WriteHeader(http.StatusOK)
	})
}

func TestHeaders_EmptyByDefault(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if h := propagate.Headers(req.Context()); h != nil {
		t.Fatalf("expected nil, got %v", h)
	}
}

func TestForwarder_PropagatesConfiguredHeaders(t *testing.T) {
	var captured http.Header
	handler := propagate.New(captureCtxHandler(&captured), "X-Request-ID", "X-Trace-ID")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "abc-123")
	req.Header.Set("X-Trace-ID", "trace-456")
	req.Header.Set("Authorization", "Bearer secret")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := captured.Get("X-Request-ID"); got != "abc-123" {
		t.Errorf("X-Request-ID: want abc-123, got %q", got)
	}
	if got := captured.Get("X-Trace-ID"); got != "trace-456" {
		t.Errorf("X-Trace-ID: want trace-456, got %q", got)
	}
	if got := captured.Get("Authorization"); got != "" {
		t.Errorf("Authorization should not be propagated, got %q", got)
	}
}

func TestForwarder_IgnoresMissingHeaders(t *testing.T) {
	var captured http.Header
	handler := propagate.New(captureCtxHandler(&captured), "X-Request-ID")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := captured.Get("X-Request-ID"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestApply_AttachesHeadersToOutboundRequest(t *testing.T) {
	var captured http.Header
	handler := propagate.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		outbound, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://example.com", nil)
		propagate.Apply(outbound, r.Context())
		captured = outbound.Header
		w.WriteHeader(http.StatusOK)
	}), "X-Correlation-ID")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-789")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got := captured.Get("X-Correlation-ID"); got != "corr-789" {
		t.Errorf("X-Correlation-ID: want corr-789, got %q", got)
	}
}

func TestForwarder_CanonicalisesHeaderNames(t *testing.T) {
	var captured http.Header
	// Register with non-canonical casing
	handler := propagate.New(captureCtxHandler(&captured), "x-request-id")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "canonical-test")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got := captured.Get("X-Request-Id"); got != "canonical-test" {
		t.Errorf("canonical header: want canonical-test, got %q", got)
	}
}
