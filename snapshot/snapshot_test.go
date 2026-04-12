package snapshot

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_DefaultCapacity(t *testing.T) {
	s := New(0)
	if s.capacity != 100 {
		t.Fatalf("expected capacity 100, got %d", s.capacity)
	}
}

func TestLen_StartsAtZero(t *testing.T) {
	s := New(10)
	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
}

func TestMiddleware_RecordsEntry(t *testing.T) {
	s := New(10)
	h := s.Middleware(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Test", "yes")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if s.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", s.Len())
	}
	e := s.Entries()[0]
	if e.Path != "/ping" {
		t.Errorf("expected path /ping, got %s", e.Path)
	}
	if e.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", e.Method)
	}
	if e.Status != http.StatusOK {
		t.Errorf("expected 200, got %d", e.Status)
	}
	if e.Headers["X-Test"] != "yes" {
		t.Errorf("expected header X-Test=yes")
	}
}

func TestMiddleware_EvictsOldestWhenFull(t *testing.T) {
	s := New(3)
	h := s.Middleware(http.HandlerFunc(okHandler))

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
	if s.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", s.Len())
	}
}

func TestMiddleware_CapturesNon200Status(t *testing.T) {
	s := New(10)
	errH := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	h := s.Middleware(errH)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/fail", nil))

	e := s.Entries()[0]
	if e.Status != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", e.Status)
	}
}

func TestHandler_ServesJSON(t *testing.T) {
	s := New(10)
	h := s.Middleware(http.HandlerFunc(okHandler))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/snapshots", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}
