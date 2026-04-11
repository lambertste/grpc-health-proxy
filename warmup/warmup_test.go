package warmup

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func okHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

func TestNew_UsesDefaultDeadlineWhenZero(t *testing.T) {
	w := &Warmup{deadline: 0}
	if w.deadline != 0 {
		t.Fatal("expected zero deadline before construction")
	}
	w2 := New(0)
	if w2.deadline != 5*time.Second {
		t.Fatalf("expected 5s default, got %v", w2.deadline)
	}
	w2.MarkReady() // prevent goroutine leak in test
}

func TestIsReady_FalseBeforeMarkReady(t *testing.T) {
	w := &Warmup{now: time.Now}
	if w.IsReady() {
		t.Fatal("expected not ready before MarkReady")
	}
}

func TestMarkReady_SetsReady(t *testing.T) {
	w := &Warmup{now: time.Now}
	w.MarkReady()
	if !w.IsReady() {
		t.Fatal("expected ready after MarkReady")
	}
}

func TestMiddleware_Returns503WhenNotReady(t *testing.T) {
	w := &Warmup{now: time.Now}
	h := w.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMiddleware_PassesThroughWhenReady(t *testing.T) {
	w := &Warmup{now: time.Now}
	w.MarkReady()
	h := w.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNew_AutomaticallyBecomesReady(t *testing.T) {
	w := New(20 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	if !w.IsReady() {
		t.Fatal("expected service to be ready after deadline")
	}
}

func TestMiddleware_BodyContainsMessage(t *testing.T) {
	w := &Warmup{now: time.Now}
	h := w.Middleware(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if body := rec.Body.String(); body == "" {
		t.Fatal("expected non-empty body on 503")
	}
}
