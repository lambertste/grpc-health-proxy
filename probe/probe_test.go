package probe

import (
	"context"
	"testing"
	"time"
)

func okProbe(_ context.Context) Result  { return Result{OK: true, Message: "ok"} }
func badProbe(_ context.Context) Result { return Result{OK: false, Message: "down"} }

func TestNew_UsesDefaultTimeoutWhenZero(t *testing.T) {
	r := New(0)
	if r.timeout != defaultTimeout {
		t.Fatalf("expected %v, got %v", defaultTimeout, r.timeout)
	}
}

func TestNew_HonorsExplicitTimeout(t *testing.T) {
	r := New(2 * time.Second)
	if r.timeout != 2*time.Second {
		t.Fatalf("expected 2s, got %v", r.timeout)
	}
}

func TestRegister_AddsProbe(t *testing.T) {
	r := New(0)
	r.Register("db", okProbe)
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.probes["db"]; !ok {
		t.Fatal("expected probe to be registered")
	}
}

func TestRunAll_ReturnsResultForEachProbe(t *testing.T) {
	r := New(0)
	r.Register("a", okProbe)
	r.Register("b", badProbe)
	res := r.RunAll(context.Background())
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	if !res["a"].OK {
		t.Error("expected probe a to be OK")
	}
	if res["b"].OK {
		t.Error("expected probe b to fail")
	}
}

func TestRunAll_PopulatesLatency(t *testing.T) {
	r := New(0)
	r.Register("slow", func(ctx context.Context) Result {
		time.Sleep(10 * time.Millisecond)
		return Result{OK: true}
	})
	res := r.RunAll(context.Background())
	if res["slow"].Latency < 10*time.Millisecond {
		t.Errorf("expected latency >= 10ms, got %v", res["slow"].Latency)
	}
}

func TestHealthy_TrueWhenAllPass(t *testing.T) {
	r := New(0)
	r.Register("x", okProbe)
	r.Register("y", okProbe)
	ok, err := r.Healthy(context.Background())
	if !ok || err != nil {
		t.Fatalf("expected healthy, got ok=%v err=%v", ok, err)
	}
}

func TestHealthy_FalseWhenAnyFails(t *testing.T) {
	r := New(0)
	r.Register("ok", okProbe)
	r.Register("bad", badProbe)
	ok, err := r.Healthy(context.Background())
	if ok {
		t.Fatal("expected unhealthy")
	}
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestHealthy_TrueWhenNoProbesRegistered(t *testing.T) {
	r := New(0)
	ok, err := r.Healthy(context.Background())
	if !ok || err != nil {
		t.Fatalf("expected healthy with no probes, got ok=%v err=%v", ok, err)
	}
}
