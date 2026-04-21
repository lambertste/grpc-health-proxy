package revision_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salrashid123/grpc-health-proxy/revision"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnNilNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil next handler")
		}
	}()
	revision.New(nil, "", "abc123")
}

func TestNew_UsesDefaultHeaderWhenEmpty(t *testing.T) {
	s := revision.New(http.HandlerFunc(okHandler), "", "abc")
	if got := s.Header(); got != "X-Revision" {
		t.Fatalf("expected X-Revision, got %q", got)
	}
}

func TestNew_HonorsExplicitHeader(t *testing.T) {
	s := revision.New(http.HandlerFunc(okHandler), "X-Build-SHA", "abc")
	if got := s.Header(); got != "X-Build-SHA" {
		t.Fatalf("expected X-Build-SHA, got %q", got)
	}
}

func TestServeHTTP_SetsRevisionHeader(t *testing.T) {
	const rev = "deadbeef"
	s := revision.New(http.HandlerFunc(okHandler), "", rev)

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Revision"); got != rev {
		t.Fatalf("expected %q, got %q", rev, got)
	}
}

func TestServeHTTP_NoHeaderWhenRevisionEmpty(t *testing.T) {
	s := revision.New(http.HandlerFunc(okHandler), "", "")

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Revision"); got != "" {
		t.Fatalf("expected empty header, got %q", got)
	}
}

func TestServeHTTP_NoHeaderWhenRevisionWhitespace(t *testing.T) {
	s := revision.New(http.HandlerFunc(okHandler), "", "   ")

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Revision"); got != "" {
		t.Fatalf("expected empty header, got %q", got)
	}
}

func TestServeHTTP_PassesThroughToNext(t *testing.T) {
	const code = http.StatusTeapot
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(code)
	})
	s := revision.New(next, "", "v1.2.3")

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != code {
		t.Fatalf("expected status %d, got %d", code, rec.Code)
	}
}

func TestRevision_ReturnsConfiguredValue(t *testing.T) {
	const rev = "v2.0.0-rc1"
	s := revision.New(http.HandlerFunc(okHandler), "", rev)
	if got := s.Revision(); got != rev {
		t.Fatalf("expected %q, got %q", rev, got)
	}
}
