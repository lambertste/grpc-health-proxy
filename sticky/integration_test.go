package sticky

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestServeHTTP_ConcurrentSafe(t *testing.T) {
	s := New(HeaderExtractor("X-Session"),
		backendHandler(0), backendHandler(1), backendHandler(2))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Session", fmt.Sprintf("user-%d", id%5))
			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)
		}(i)
	}
	wg.Wait()
}

func TestServeHTTP_SingleBackendAlwaysReceivesRequest(t *testing.T) {
	s := New(HeaderExtractor("X-Session"), backendHandler(42))

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Session", fmt.Sprintf("key-%d", i))
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, req)
		if rec.Body.String() != "backend-42" {
			t.Fatalf("expected backend-42, got %s", rec.Body.String())
		}
	}
}
