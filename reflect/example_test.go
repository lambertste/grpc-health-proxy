package reflect_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/salrashid123/grpc-health-proxy/reflect"
)

// ExampleNew demonstrates wrapping a handler so that request metadata is
// echoed back in the response headers.
func ExampleNew() {
	backend := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rfl := reflect.New(backend)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Proto = "HTTP/1.1"
	req.RemoteAddr = "192.168.1.10:54321"
	rec := httptest.NewRecorder()

	rfl.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("X-Reflect-Method"))
	fmt.Println(rec.Header().Get("X-Reflect-Path"))
	// Output:
	// GET
	// /healthz
}
