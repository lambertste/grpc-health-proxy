package shadow_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/salmanahmad/grpc-health-proxy/shadow"
)

// recordingDoer captures the last mirrored request.
type recordingDoer struct {
	mu  sync.Mutex
	reqs []*http.Request
}

func (d *recordingDoer) Do(r *http.Request) (*http.Response, error) {
	d.mu.Lock()
	d.reqs = append(d.reqs, r)
	d.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func (d *recordingDoer) last() *http.Request {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.reqs) == 0 {
		return nil
	}
	return d.reqs[len(d.reqs)-1]
}

func TestNew_UsesDefaultTimeoutWhenZero(t *testing.T) {
	s := shadow.New("http://shadow", nil, 0)
	if s == nil {
		t.Fatal("expected non-nil Shadow")
	}
}

func TestMiddleware_PassesThroughToPrimary(t *testing.T) {
	doer := &recordingDoer{}
	s := shadow.New("http://shadow", doer, 100*time.Millisecond)

	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	s.Middleware(primary).ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418 from primary, got %d", rec.Code)
	}
}

func TestMiddleware_MirrorsRequestToShadow(t *testing.T) {
	doer := &recordingDoer{}
	s := shadow.New("http://shadow", doer, 200*time.Millisecond)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/check", strings.NewReader("ping"))
	req.Header.Set("X-Custom", "value")
	s.Middleware(next).ServeHTTP(httptest.NewRecorder(), req)

	// give goroutine time to run
	time.Sleep(50 * time.Millisecond)

	last := doer.last()
	if last == nil {
		t.Fatal("expected shadow request to be sent")
	}
	if last.Header.Get("X-Custom") != "value" {
		t.Errorf("expected header to be forwarded, got %q", last.Header.Get("X-Custom"))
	}
	if last.Method != http.MethodPost {
		t.Errorf("expected POST, got %s", last.Method)
	}
}
