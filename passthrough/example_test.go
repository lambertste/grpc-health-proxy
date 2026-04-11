package passthrough_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourorg/grpc-health-proxy/passthrough"
)

// ExampleNew demonstrates routing OPTIONS preflight requests to a dedicated
// CORS handler while all other traffic continues through the primary handler.
func ExampleNew() {
	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "primary")
	})

	cors := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusNoContent)
	})

	h := passthrough.New(primary, passthrough.MethodIn(http.MethodOptions), cors)

	// Preflight request — served by cors handler.
	pre := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, pre)
	fmt.Println("preflight:", w.Code)

	// Normal request — served by primary handler.
	norm := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, norm)
	fmt.Println("normal:", w2.Code)

	// Output:
	// preflight: 204
	// normal: 200
}
