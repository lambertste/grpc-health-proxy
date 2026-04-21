package pacing

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew_UsesDefaultIntervalWhenZero(t *testing.T) {
	p := New(0)
	if p.Interval() != 10*time.Millisecond {
		t.Fatalf("expected 10ms default, got %v", p.Interval())
	}
}

func TestNew_HonorsExplicitInterval(t *testing.T) {
	p := New(50 * time.Millisecond)
	if p.Interval() != 50*time.Millisecond {
		t.Fatalf("expected 50ms, got %v", p.Interval())
	}
}

func TestWait_NoDelayOnFirstCall(t *testing.T) {
	p := New(100 * time.Millisecond)
	var slept time.Duration
	p.sleep = func(d time.Duration) { slept = d }
	p.clock = time.Now

	p.Wait()
	if slept != 0 {
		t.Fatalf("expected no sleep on first call, got %v", slept)
	}
}

func TestWait_DelaysSubsequentCallsWithinInterval(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	p := New(100 * time.Millisecond)
	p.clock = func() time.Time { return now }
	var slept time.Duration
	p.sleep = func(d time.Duration) { slept = d }

	p.Wait() // first call — no delay
	p.Wait() // second call — should sleep ~100 ms

	if slept != 100*time.Millisecond {
		t.Fatalf("expected 100ms sleep, got %v", slept)
	}
}

func TestWait_NoDelayWhenIntervalElapsed(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	calls := 0
	p := New(50 * time.Millisecond)
	p.clock = func() time.Time {
		calls++
		if calls == 1 {
			return base
		}
		return base.Add(200 * time.Millisecond) // well past interval
	}
	var slept time.Duration
	p.sleep = func(d time.Duration) { slept = d }

	p.Wait()
	p.Wait()

	if slept != 0 {
		t.Fatalf("expected no sleep when interval already elapsed, got %v", slept)
	}
}

func TestMiddleware_ForwardsRequest(t *testing.T) {
	p := New(0)
	p.sleep = func(time.Duration) {} // no-op

	var called int32
	h := p.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&called, 1)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatal("expected handler to be called once")
	}
}
