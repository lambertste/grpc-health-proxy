package tags_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"grpc-health-proxy/tags"
)

func TestNew_EmptyOnCreate(t *testing.T) {
	tg := tags.New()
	if got := tg.All(); len(got) != 0 {
		t.Fatalf("expected empty tags, got %v", got)
	}
}

func TestSet_StoresValue(t *testing.T) {
	tg := tags.New()
	tg.Set("env", "prod")
	v, ok := tg.Get("env")
	if !ok || v != "prod" {
		t.Fatalf("expected 'prod', got %q ok=%v", v, ok)
	}
}

func TestGet_MissingKeyReturnsFalse(t *testing.T) {
	tg := tags.New()
	_, ok := tg.Get("missing")
	if ok {
		t.Fatal("expected ok=false for missing key")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	tg := tags.New()
	tg.Set("a", "1")
	tg.Set("b", "2")
	all := tg.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating copy must not affect original.
	all["c"] = "3"
	if _, ok := tg.Get("c"); ok {
		t.Fatal("mutation of copy affected original")
	}
}

func TestFromContext_NilWhenAbsent(t *testing.T) {
	if got := tags.FromContext(context.Background()); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestContextWithTags_RoundTrip(t *testing.T) {
	tg := tags.New()
	tg.Set("x", "y")
	ctx := tags.ContextWithTags(context.Background(), tg)
	got := tags.FromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil Tags from context")
	}
	if v, _ := got.Get("x"); v != "y" {
		t.Fatalf("expected 'y', got %q", v)
	}
}

func TestMiddleware_InjectsTags(t *testing.T) {
	var captured *tags.Tags
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = tags.FromContext(r.Context())
		if captured != nil {
			captured.Set("injected", "true")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	tags.Middleware(handler).ServeHTTP(rec, req)

	if captured == nil {
		t.Fatal("expected Tags in context, got nil")
	}
	if v, ok := captured.Get("injected"); !ok || v != "true" {
		t.Fatalf("expected 'true', got %q ok=%v", v, ok)
	}
}

func TestMiddleware_FreshTagsPerRequest(t *testing.T) {
	var first, second *tags.Tags
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tg := tags.FromContext(r.Context())
		if first == nil {
			first = tg
		} else {
			second = tg
		}
		w.WriteHeader(http.StatusOK)
	})
	mw := tags.Middleware(handler)
	for i := 0; i < 2; i++ {
		mw.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if first == second {
		t.Fatal("expected distinct Tags instances per request")
	}
}
