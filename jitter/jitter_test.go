package jitter_test

import (
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/jitter"
)

func TestFull_ZeroInputReturnsZero(t *testing.T) {
	if got := jitter.Full(0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestFull_NegativeInputReturnsZero(t *testing.T) {
	if got := jitter.Full(-time.Second); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestFull_ResultWithinRange(t *testing.T) {
	d := 100 * time.Millisecond
	for i := 0; i < 1000; i++ {
		got := jitter.Full(d)
		if got < 0 || got >= d {
			t.Fatalf("Full(%v) = %v; want [0, %v)", d, got, d)
		}
	}
}

func TestEqual_ZeroInputReturnsZero(t *testing.T) {
	if got := jitter.Equal(0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestEqual_ResultWithinHalfToFull(t *testing.T) {
	d := 200 * time.Millisecond
	half := d / 2
	for i := 0; i < 1000; i++ {
		got := jitter.Equal(d)
		if got < half || got >= d {
			t.Fatalf("Equal(%v) = %v; want [%v, %v)", d, got, half, d)
		}
	}
}

func TestDecorrelated_ZeroBaseReturnsZero(t *testing.T) {
	if got := jitter.Decorrelated(0, time.Second, 10*time.Second); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestDecorrelated_ResultCappedAtCap(t *testing.T) {
	base := 100 * time.Millisecond
	cap_ := 500 * time.Millisecond
	prev := 10 * time.Second // large prev to push result high
	for i := 0; i < 500; i++ {
		got := jitter.Decorrelated(base, prev, cap_)
		if got > cap_ {
			t.Fatalf("Decorrelated exceeded cap: got %v, cap %v", got, cap_)
		}
	}
}

func TestDecorrelated_ResultAtLeastBase(t *testing.T) {
	base := 50 * time.Millisecond
	for i := 0; i < 500; i++ {
		got := jitter.Decorrelated(base, base, time.Second)
		if got < base {
			t.Fatalf("Decorrelated below base: got %v, base %v", got, base)
		}
	}
}
