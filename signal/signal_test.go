package signal

import (
	"context"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestNew_UsesDefaultTimeoutWhenZero(t *testing.T) {
	h := New(0)
	if h.timeout != 30*time.Second {
		t.Fatalf("expected 30s default timeout, got %v", h.timeout)
	}
}

func TestNew_HonorsExplicitTimeout(t *testing.T) {
	h := New(5 * time.Second)
	if h.timeout != 5*time.Second {
		t.Fatalf("expected 5s timeout, got %v", h.timeout)
	}
}

func TestOnShutdown_RegistersListener(t *testing.T) {
	h := New(time.Second)
	h.OnShutdown(func() {})
	h.OnShutdown(func() {})
	if len(h.listeners) != 2 {
		t.Fatalf("expected 2 listeners, got %d", len(h.listeners))
	}
}

func TestNotify_CallsAllListeners(t *testing.T) {
	h := New(time.Second)

	var called int32
	h.OnShutdown(func() { atomic.AddInt32(&called, 1) })
	h.OnShutdown(func() { atomic.AddInt32(&called, 1) })
	h.OnShutdown(func() { atomic.AddInt32(&called, 1) })

	h.notify()

	if atomic.LoadInt32(&called) != 3 {
		t.Fatalf("expected 3 listeners called, got %d", called)
	}
}

func TestNotify_TimesOutSlowListeners(t *testing.T) {
	h := New(50 * time.Millisecond)

	var fast int32
	h.OnShutdown(func() { atomic.AddInt32(&fast, 1) })
	h.OnShutdown(func() { time.Sleep(10 * time.Second) }) // intentionally slow

	start := time.Now()
	h.notify()
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Fatalf("notify blocked too long: %v", elapsed)
	}
	if atomic.LoadInt32(&fast) != 1 {
		t.Fatal("fast listener was not called")
	}
}

func TestWait_CancelledParentTriggersShutdown(t *testing.T) {
	h := New(time.Second)

	var called int32
	h.OnShutdown(func() { atomic.AddInt32(&called, 1) })

	parent, cancel := context.WithCancel(context.Background())
	ctx := h.Wait(parent)

	cancel() // simulate external cancellation instead of OS signal

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("context was not cancelled after parent cancellation")
	}

	time.Sleep(50 * time.Millisecond) // allow listener goroutine to run
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected listener to be called once, got %d", called)
	}
}

func TestWait_ContextCancelledOnSignal(t *testing.T) {
	h := New(time.Second)
	ctx := h.Wait(context.Background())

	// send signal to self
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM) //nolint:errcheck

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("context was not cancelled after SIGTERM")
	}
}
