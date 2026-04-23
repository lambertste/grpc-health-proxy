package reflect_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/salrashid123/grpc-health-proxy/reflect"
)

func TestServeHTTP_ConcurrentSafe(t *testing.T) {
	rfl := reflect.New(http.HandlerFunc(okHandler))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			rfl.ServeHTTP(rec, req)
			if rec.Header().Get("X-Reflect-Method") != "GET" {
				t.Errorf("goroutine %d: unexpected method header", i)
			}
		}(i)
	}
	wg.Wait()
}

func TestServeHTTP_AllFieldsPresent(t *testing.T) {
	rfl := reflect.New(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/resource", nil)
	req.Proto = "HTTP/2.0"
	req.RemoteAddr = "172.16.0.5:8080"
	rec := httptest.NewRecorder()

	rfl.ServeHTTP(rec, req)

	cases := map[string]string{
		"X-Reflect-Method": "PUT",
		"X-Reflect-Path":   "/api/v1/resource",
		"X-Reflect-Remote": "172.16.0.5:8080",
		"X-Reflect-Proto":  "HTTP/2.0",
	}
	for header, want := range cases {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s: want %q, got %q", header, want, got)
		}
	}
}
