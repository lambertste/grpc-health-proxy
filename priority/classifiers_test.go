package priority

import (
	"net/http/httptest"
	"testing"
)

func TestHeaderClassifier_ReturnsLow(t *testing.T) {
	c := HeaderClassifier("X-Priority")
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Priority", "low")
	if got := c(r); got != Low {
		t.Fatalf("expected Low, got %v", got)
	}
}

func TestHeaderClassifier_ReturnsHigh(t *testing.T) {
	c := HeaderClassifier("X-Priority")
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Priority", "HIGH")
	if got := c(r); got != High {
		t.Fatalf("expected High, got %v", got)
	}
}

func TestHeaderClassifier_DefaultsToNormal(t *testing.T) {
	c := HeaderClassifier("X-Priority")
	r := httptest.NewRequest("GET", "/", nil)
	if got := c(r); got != Normal {
		t.Fatalf("expected Normal, got %v", got)
	}
}

func TestPathPrefixClassifier_High(t *testing.T) {
	c := PathPrefixClassifier("/critical", "/background")
	r := httptest.NewRequest("GET", "/critical/health", nil)
	if got := c(r); got != High {
		t.Fatalf("expected High, got %v", got)
	}
}

func TestPathPrefixClassifier_Low(t *testing.T) {
	c := PathPrefixClassifier("/critical", "/background")
	r := httptest.NewRequest("GET", "/background/sync", nil)
	if got := c(r); got != Low {
		t.Fatalf("expected Low, got %v", got)
	}
}

func TestPathPrefixClassifier_Normal(t *testing.T) {
	c := PathPrefixClassifier("/critical", "/background")
	r := httptest.NewRequest("GET", "/api/v1/data", nil)
	if got := c(r); got != Normal {
		t.Fatalf("expected Normal, got %v", got)
	}
}

func TestChainClassifier_FirstNonNormalWins(t *testing.T) {
	header := HeaderClassifier("X-Priority")
	path := PathPrefixClassifier("/critical", "")
	c := ChainClassifier(header, path)

	r := httptest.NewRequest("GET", "/critical/x", nil)
	if got := c(r); got != High {
		t.Fatalf("expected High from path, got %v", got)
	}

	r2 := httptest.NewRequest("GET", "/critical/x", nil)
	r2.Header.Set("X-Priority", "low")
	if got := c(r2); got != Low {
		t.Fatalf("expected Low from header (checked first), got %v", got)
	}
}

func TestChainClassifier_FallsBackToNormal(t *testing.T) {
	c := ChainClassifier(HeaderClassifier("X-Priority"), PathPrefixClassifier("/a", "/b"))
	r := httptest.NewRequest("GET", "/other", nil)
	if got := c(r); got != Normal {
		t.Fatalf("expected Normal, got %v", got)
	}
}
