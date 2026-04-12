package checkpoint_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/grpc-health-proxy/checkpoint"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_DefaultCapacity(t *testing.T) {
	l := checkpoint.New(0)
	if l == nil {
		t.Fatal("expected non-nil Log")
	}
}

func TestLen_StartsAtZero(t *testing.T) {
	l := checkpoint.New(10)
	if l.Len() != 0 {
		t.Fatalf("expected 0, got %d", l.Len())
	}
}

func TestMiddleware_RecordsEntry(t *testing.T) {
	l := checkpoint.New(10)
	h := l.Middleware(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if l.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", l.Len())
	}
	entries := l.Entries()
	if entries[0].Method != http.MethodGet {
		t.Errorf("expected GET, got %s", entries[0].Method)
	}
	if entries[0].Path != "/health" {
		t.Errorf("expected /health, got %s", entries[0].Path)
	}
	if entries[0].Status != http.StatusOK {
		t.Errorf("expected 200, got %d", entries[0].Status)
	}
}

func TestMiddleware_CapturesNon200Status(t *testing.T) {
	l := checkpoint.New(10)
	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	req := httptest.NewRequest(http.MethodPost, "/check", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if l.Entries()[0].Status != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", l.Entries()[0].Status)
	}
}

func TestLog_EvictsOldestWhenFull(t *testing.T) {
	cap := 3
	l := checkpoint.New(cap)
	h := l.Middleware(http.HandlerFunc(okHandler))

	for i := 0; i < cap+2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}

	if l.Len() != cap {
		t.Fatalf("expected %d entries after eviction, got %d", cap, l.Len())
	}
}

func TestEntries_ReturnsCopy(t *testing.T) {
	l := checkpoint.New(10)
	h := l.Middleware(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	a := l.Entries()
	a[0].Path = "/mutated"
	b := l.Entries()
	if b[0].Path == "/mutated" {
		t.Error("Entries should return an independent copy")
	}
}

func TestEntry_String(t *testing.T) {
	e := checkpoint.Entry{
		Timestamp: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		Method:    "GET",
		Path:      "/ping",
		Status:    200,
	}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
