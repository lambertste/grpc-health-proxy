package window

import (
	"testing"
	"time"
)

func TestNew_PanicsOnInvalidSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for size=0")
		}
	}()
	New(time.Second, 0)
}

func TestNew_PanicsOnNonPositiveWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for window=0")
		}
	}()
	New(0, 10)
}

func TestCount_StartsAtZero(t *testing.T) {
	c := New(time.Second, 10)
	if got := c.Count(); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestAdd_IncrementsCount(t *testing.T) {
	c := New(time.Second, 10)
	c.Add(3)
	c.Add(2)
	if got := c.Count(); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestCount_ExcludesExpiredBuckets(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	c := New(time.Second, 10) // 10 buckets of 100ms each
	c.now = func() time.Time { return base }
	c.Add(7)

	// advance past the full window — all buckets are now stale
	c.now = func() time.Time { return base.Add(2 * time.Second) }
	if got := c.Count(); got != 0 {
		t.Fatalf("expected 0 after window expired, got %d", got)
	}
}

func TestCount_RetainsRecentBuckets(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	c := New(time.Second, 10) // buckets of 100ms
	c.now = func() time.Time { return base }
	c.Add(4)

	// advance by half the window — event should still be visible
	c.now = func() time.Time { return base.Add(500 * time.Millisecond) }
	if got := c.Count(); got != 4 {
		t.Fatalf("expected 4, got %d", got)
	}
}

func TestReset_ClearsAll(t *testing.T) {
	c := New(time.Second, 10)
	c.Add(10)
	c.Reset()
	if got := c.Count(); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestAdd_ConcurrentSafe(t *testing.T) {
	c := New(time.Second, 10)
	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			c.Add(1)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	if got := c.Count(); got != 100 {
		t.Fatalf("expected 100, got %d", got)
	}
}
