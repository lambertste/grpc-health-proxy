package replay

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestConcurrentRecord_Safe(t *testing.T) {
	rec := New(100)
	h := rec.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			h.ServeHTTP(httptest.NewRecorder(), req)
		}()
	}
	wg.Wait()

	if rec.Len() > 100 {
		t.Fatalf("len %d exceeds cap 100", rec.Len())
	}
}

func TestReplay_ConcurrentSafe(t *testing.T) {
	rec := New(20)
	h := rec.Middleware(http.HandlerFunc(okHandler))
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec.Replay(http.HandlerFunc(okHandler))
		}()
	}
	wg.Wait()
}
