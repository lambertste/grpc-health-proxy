package deadletter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpc-health-proxy/deadletter"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func errHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestNew_DefaultCapacity(t *testing.T) {
	q := deadletter.New(0)
	if q == nil {
		t.Fatal("expected non-nil queue")
	}
}

func TestLen_StartsAtZero(t *testing.T) {
	q := deadletter.New(10)
	if q.Len() != 0 {
		t.Fatalf("expected 0, got %d", q.Len())
	}
}

func TestMiddleware_DoesNotRecordSuccess(t *testing.T) {
	q := deadletter.New(10)
	h := q.Middleware(deadletter.IsError, http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if q.Len() != 0 {
		t.Fatalf("expected 0 entries, got %d", q.Len())
	}
}

func TestMiddleware_RecordsError(t *testing.T) {
	q := deadletter.New(10)
	h := q.Middleware(deadletter.IsError, http.HandlerFunc(errHandler))

	req := httptest.NewRequest(http.MethodPost, "/fail", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if q.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", q.Len())
	}
	e := q.Entries()[0]
	if e.Method != http.MethodPost {
		t.Errorf("expected POST, got %s", e.Method)
	}
	if e.Path != "/fail" {
		t.Errorf("expected /fail, got %s", e.Path)
	}
	if e.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", e.StatusCode)
	}
}

func TestQueue_EvictsOldestWhenFull(t *testing.T) {
	q := deadletter.New(3)
	h := q.Middleware(deadletter.IsError, http.HandlerFunc(errHandler))

	paths := []string{"/a", "/b", "/c", "/d"}
	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}

	if q.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", q.Len())
	}
	if q.Entries()[0].Path != "/b" {
		t.Errorf("expected oldest entry to be /b, got %s", q.Entries()[0].Path)
	}
}

func TestEntries_ReturnsCopy(t *testing.T) {
	q := deadletter.New(10)
	h := q.Middleware(deadletter.IsError, http.HandlerFunc(errHandler))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	a := q.Entries()
	a[0].Path = "/mutated"
	b := q.Entries()
	if b[0].Path == "/mutated" {
		t.Error("Entries should return a copy, not a reference")
	}
}
