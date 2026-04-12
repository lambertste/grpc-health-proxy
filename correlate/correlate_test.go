package correlate_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/grpc-health-proxy/correlate"
)

func TestIDFromContext_EmptyByDefault(t *testing.T) {
	if got := correlate.IDFromContext(context.Background()); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestContextWithID_RoundTrip(t *testing.T) {
	ctx := correlate.ContextWithID(context.Background(), "abc-123")
	if got := correlate.IDFromContext(ctx); got != "abc-123" {
		t.Fatalf("expected abc-123, got %q", got)
	}
}

func TestMiddleware_PropagatesExistingHeader(t *testing.T) {
	const wantID = "existing-id-42"
	var gotID string

	handler := correlate.Middleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = correlate.IDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(correlate.DefaultHeader, wantID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotID != wantID {
		t.Fatalf("context ID: want %q, got %q", wantID, gotID)
	}
	if rec.Header().Get(correlate.DefaultHeader) != wantID {
		t.Fatalf("response header: want %q, got %q", wantID, rec.Header().Get(correlate.DefaultHeader))
	}
}

func TestMiddleware_GeneratesIDWhenAbsent(t *testing.T) {
	var gotID string

	handler := correlate.Middleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = correlate.IDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotID == "" {
		t.Fatal("expected a generated correlation ID, got empty string")
	}
	if rec.Header().Get(correlate.DefaultHeader) != gotID {
		t.Fatalf("response header mismatch: want %q, got %q", gotID, rec.Header().Get(correlate.DefaultHeader))
	}
}

func TestMiddleware_UniqueIDsPerRequest(t *testing.T) {
	ids := make([]string, 3)

	handler := correlate.Middleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := range ids {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		ids[i] = rec.Header().Get(correlate.DefaultHeader)
	}

	for i := 1; i < len(ids); i++ {
		if ids[i] == ids[i-1] {
			t.Fatalf("expected unique IDs but got duplicate: %q", ids[i])
		}
	}
}

func TestMiddleware_CustomHeader(t *testing.T) {
	const customHeader = "X-Request-ID"
	const wantID = "custom-99"
	var gotID string

	handler := correlate.Middleware(customHeader)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = correlate.IDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(customHeader, wantID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotID != wantID {
		t.Fatalf("want %q, got %q", wantID, gotID)
	}
}
