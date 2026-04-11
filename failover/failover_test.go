package failover_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourorg/grpc-health-proxy/failover"
)

func TestNew_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty targets")
		}
	}()
	failover.New(nil)
}

func TestDo_SucceedsOnFirstTarget(t *testing.T) {
	pool := failover.New([]string{"host-a:443", "host-b:443"})
	called := map[string]int{}
	target, err := pool.Do(context.Background(), func(_ context.Context, addr string) error {
		called[addr]++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == "" {
		t.Fatal("expected non-empty target")
	}
	if len(called) != 1 {
		t.Fatalf("expected exactly one call, got %d", len(called))
	}
}

func TestDo_FailsOverToSecondTarget(t *testing.T) {
	pool := failover.New([]string{"bad:443", "good:443"})
	var order []string
	target, err := pool.Do(context.Background(), func(_ context.Context, addr string) error {
		order = append(order, addr)
		if addr == "bad:443" {
			return errors.New("unhealthy")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target != "good:443" {
		t.Fatalf("expected good:443, got %s", target)
	}
	if len(order) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(order))
	}
}

func TestDo_ReturnsErrAllUnhealthy(t *testing.T) {
	pool := failover.New([]string{"a:443", "b:443"})
	_, err := pool.Do(context.Background(), func(_ context.Context, _ string) error {
		return errors.New("down")
	})
	if !errors.Is(err, failover.ErrAllUnhealthy) {
		t.Fatalf("expected ErrAllUnhealthy, got %v", err)
	}
}

func TestTargets_ReturnsCopy(t *testing.T) {
	orig := []string{"x:443"}
	pool := failover.New(orig)
	copy1 := pool.Targets()
	copy1[0] = "mutated"
	if pool.Targets()[0] != "x:443" {
		t.Fatal("Targets() should return an independent copy")
	}
}

func TestDo_RoundRobinStartOffset(t *testing.T) {
	pool := failover.New([]string{"a:443", "b:443"})
	first := map[string]int{}
	for i := 0; i < 4; i++ {
		_, _ = pool.Do(context.Background(), func(_ context.Context, addr string) error {
			first[addr]++
			return nil // succeed immediately so only first target is recorded
		})
	}
	if first["a:443"] == 0 || first["b:443"] == 0 {
		t.Fatal("expected both targets to be selected as the first target across calls")
	}
}
