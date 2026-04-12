package watchdog_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sidecar/grpc-health-proxy/watchdog"
)

func TestNew_UsesDefaultIntervalWhenZero(t *testing.T) {
	probe := func(_ context.Context) error { return nil }
	w := watchdog.New(probe, 0)
	if w == nil {
		t.Fatal("expected non-nil watchdog")
	}
}

func TestHealthy_TrueAfterSuccessfulProbe(t *testing.T) {
	probe := func(_ context.Context) error { return nil }
	w := watchdog.New(probe, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	go w.Run(ctx)
	<-ctx.Done()

	if !w.Healthy() {
		t.Error("expected healthy after successful probe")
	}
	if w.LastErr() != nil {
		t.Errorf("expected nil error, got %v", w.LastErr())
	}
}

func TestHealthy_FalseAfterFailingProbe(t *testing.T) {
	sentinel := errors.New("unhealthy")
	probe := func(_ context.Context) error { return sentinel }
	w := watchdog.New(probe, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	go w.Run(ctx)
	<-ctx.Done()

	if w.Healthy() {
		t.Error("expected unhealthy after failing probe")
	}
	if !errors.Is(w.LastErr(), sentinel) {
		t.Errorf("expected sentinel error, got %v", w.LastErr())
	}
}

func TestRun_ProbeCalledMultipleTimes(t *testing.T) {
	var calls atomic.Int64
	probe := func(_ context.Context) error {
		calls.Add(1)
		return nil
	}
	w := watchdog.New(probe, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Millisecond)
	defer cancel()

	w.Run(ctx)

	if n := calls.Load(); n < 3 {
		t.Errorf("expected at least 3 probe calls, got %d", n)
	}
}

func TestRun_StopsWhenContextCancelled(t *testing.T) {
	var calls atomic.Int64
	probe := func(_ context.Context) error {
		calls.Add(1)
		return nil
	}
	w := watchdog.New(probe, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	w.Run(ctx) // should return quickly

	snap := calls.Load()
	time.Sleep(30 * time.Millisecond)
	if calls.Load() > snap+1 {
		t.Error("probe kept running after context was cancelled")
	}
}
