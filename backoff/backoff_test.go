package backoff_test

import (
	"testing"
	"time"

	"github.com/yourorg/grpc-health-proxy/backoff"
)

func TestConstant_AlwaysReturnsSameDuration(t *testing.T) {
	s := backoff.Constant(200 * time.Millisecond)
	for attempt := 0; attempt < 5; attempt++ {
		if got := s(attempt); got != 200*time.Millisecond {
			t.Fatalf("attempt %d: got %v, want 200ms", attempt, got)
		}
	}
}

func TestLinear_GrowsLinearly(t *testing.T) {
	base := 100 * time.Millisecond
	s := backoff.Linear(base)
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 300 * time.Millisecond},
		{4, 500 * time.Millisecond},
	}
	for _, tc := range cases {
		if got := s(tc.attempt); got != tc.want {
			t.Errorf("attempt %d: got %v, want %v", tc.attempt, got, tc.want)
		}
	}
}

func TestExponential_DoublesEachAttempt(t *testing.T) {
	s := backoff.Exponential(100*time.Millisecond, 0)
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
	}
	for _, tc := range cases {
		if got := s(tc.attempt); got != tc.want {
			t.Errorf("attempt %d: got %v, want %v", tc.attempt, got, tc.want)
		}
	}
}

func TestExponential_CapsAtMax(t *testing.T) {
	max := 300 * time.Millisecond
	s := backoff.Exponential(100*time.Millisecond, max)
	for attempt := 2; attempt < 10; attempt++ {
		if got := s(attempt); got != max {
			t.Errorf("attempt %d: got %v, want cap %v", attempt, got, max)
		}
	}
}

func TestExponential_ZeroBaseReturnsZero(t *testing.T) {
	s := backoff.Exponential(0, time.Second)
	if got := s(3); got != 0 {
		t.Fatalf("got %v, want 0", got)
	}
}

func TestDefault_IsNonNilAndIncreasing(t *testing.T) {
	prev := backoff.Default(0)
	for attempt := 1; attempt < 6; attempt++ {
		curr := backoff.Default(attempt)
		if curr <= prev {
			// once capped, equal is acceptable
			if curr < prev {
				t.Errorf("attempt %d: duration decreased: %v -> %v", attempt, prev, curr)
			}
		}
		prev = curr
	}
}
