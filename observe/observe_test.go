package observe

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func okHandler(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	})
}

func TestNew_NilSinkDoesNotPanic(t *testing.T) {
	obs := New(okHandler(http.StatusOK), nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	obs.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServeHTTP_CapturesStatusCode(t *testing.T) {
	var got Event
	obs := New(okHandler(http.StatusCreated), func(e Event) { got = e })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	obs.ServeHTTP(rec, req)
	if got.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", got.StatusCode)
	}
}

func TestServeHTTP_CapturesMethodAndPath(t *testing.T) {
	var got Event
	obs := New(okHandler(http.StatusOK), func(e Event) { got = e })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/42", nil)
	obs.ServeHTTP(rec, req)
	if got.Method != http.MethodDelete {
		t.Fatalf("expected DELETE, got %s", got.Method)
	}
	if got.Path != "/users/42" {
		t.Fatalf("expected /users/42, got %s", got.Path)
	}
}

func TestServeHTTP_CapturesPositiveLatency(t *testing.T) {
	var got Event
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	obs := New(slow, func(e Event) { got = e })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	obs.ServeHTTP(rec, req)
	if got.Latency <= 0 {
		t.Fatalf("expected positive latency, got %v", got.Latency)
	}
}

func TestServeHTTP_DefaultStatus200WhenNotSet(t *testing.T) {
	var got Event
	silent := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// deliberately does not call WriteHeader
		_, _ = w.Write([]byte("ok"))
	})
	obs := New(silent, func(e Event) { got = e })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	obs.ServeHTTP(rec, req)
	if got.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", got.StatusCode)
	}
}
