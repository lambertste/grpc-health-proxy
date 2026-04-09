package circuit_test

import (
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/circuit"
)

func TestNew_InitialStateClosed(t *testing.T) {
	b := circuit.New(3, 50*time.Millisecond)
	if got := b.State(); got != circuit.StateClosed {
		t.Fatalf("expected StateClosed, got %v", got)
	}
}

func TestAllow_ClosedAlwaysAllows(t *testing.T) {
	b := circuit.New(3, 50*time.Millisecond)
	if !b.Allow() {
		t.Fatal("expected Allow to return true in closed state")
	}
}

func TestBreaker_OpensAfterMaxFailures(t *testing.T) {
	b := circuit.New(3, 50*time.Millisecond)
	for i := 0; i < 3; i++ {
		b.RecordFailure()
	}
	if got := b.State(); got != circuit.StateOpen {
		t.Fatalf("expected StateOpen after max failures, got %v", got)
	}
	if b.Allow() {
		t.Fatal("expected Allow to return false in open state")
	}
}

func TestBreaker_TransitionsToHalfOpenAfterTimeout(t *testing.T) {
	b := circuit.New(1, 30*time.Millisecond)
	b.RecordFailure()
	if b.State() != circuit.StateOpen {
		t.Fatal("expected StateOpen")
	}
	time.Sleep(40 * time.Millisecond)
	if !b.Allow() {
		t.Fatal("expected Allow to return true after reset timeout")
	}
	if b.State() != circuit.StateHalfOpen {
		t.Fatalf("expected StateHalfOpen, got %v", b.State())
	}
}

func TestBreaker_ClosesOnSuccessFromHalfOpen(t *testing.T) {
	b := circuit.New(1, 20*time.Millisecond)
	b.RecordFailure()
	time.Sleep(30 * time.Millisecond)
	b.Allow() // transition to half-open
	b.RecordSuccess()
	if got := b.State(); got != circuit.StateClosed {
		t.Fatalf("expected StateClosed after success, got %v", got)
	}
}

func TestBreaker_RecordSuccess_ResetsFailures(t *testing.T) {
	b := circuit.New(3, 50*time.Millisecond)
	b.RecordFailure()
	b.RecordFailure()
	b.RecordSuccess()
	// After success, two more failures should not open the breaker yet
	b.RecordFailure()
	b.RecordFailure()
	if got := b.State(); got != circuit.StateClosed {
		t.Fatalf("expected StateClosed, got %v", got)
	}
}
