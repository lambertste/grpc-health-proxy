package tracing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpc-health-proxy/tracing"
)

func TestIDFromContext_EmptyByDefault(t *testing.T) {
	if got := tracing.IDFromContext(context.Background()); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestContextWithID_RoundTrip(t *testing.T) {
	ctx := tracing.ContextWithID(context.Background(), "abc123")
	if got := tracing.IDFromContext(ctx); got != "abc123" {
		t.Fatalf("expected abc123, got %q", got)
	}
}

func TestMiddleware_PropagatesExistingHeader(t *testing.T) {
	const want = "existing-trace-id"
	var got string

	handler := tracing.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = tracing.IDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(tracing.TraceIDHeader, want)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got != want {
		t.Fatalf("context ID: want %q, got %q", want, got)
	}
	if w.Header().Get(tracing.TraceIDHeader) != want {
		t.Fatalf("response header: want %q, got %q", want, w.Header().Get(tracing.TraceIDHeader))
	}
}

func TestMiddleware_GeneratesIDWhenAbsent(t *testing.T) {
	var got string

	handler := tracing.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = tracing.IDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got == "" {
		t.Fatal("expected a generated trace ID, got empty string")
	}
	if w.Header().Get(tracing.TraceIDHeader) != got {
		t.Fatalf("response header should match context ID")
	}
}

func TestMiddleware_UniqueIDsPerRequest(t *testing.T) {
	ids := make(map[string]struct{})

	handler := tracing.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ids[tracing.IDFromContext(r.Context())] = struct{}{}
	}))

	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	if len(ids) != 20 {
		t.Fatalf("expected 20 unique IDs, got %d", len(ids))
	}
}
