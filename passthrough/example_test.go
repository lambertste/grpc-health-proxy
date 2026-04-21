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

// ExampleNew_multipleConditions demonstrates chaining multiple bypass handlers
// so that both OPTIONS and HEAD requests are handled by dedicated handlers.
func ExampleNew_multipleConditions() {
	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cors := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	health := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Chain: OPTIONS -> cors, HEAD -> health, everything else -> primary.
	h := passthrough.New(
		passthrough.New(primary, passthrough.MethodIn(http.MethodHead), health),
		passthrough.MethodIn(http.MethodOptions),
		cors,
	)

	opts := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, opts)
	fmt.Println("options:", w.Code)

	head := httptest.NewRequest(http.MethodHead, "/", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, head)
	fmt.Println("head:", w2.Code)

	// Output:
	// options: 204
	// head: 200
}

// ExampleNew_pathPrefix demonstrates bypassing the primary handler for
// requests whose path matches a specific prefix, such as a health-check route.
func ExampleNew_pathPrefix() {
	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "app")
	})

	healthz := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "healthy")
	})

	// Route /healthz to the dedicated healthz handler; everything else to primary.
	isHealthz := passthrough.ConditionFunc(func(r *http.Request) bool {
		return r.URL.Path == "/healthz"
	})
	h := passthrough.New(primary, isHealthz, healthz)

	hReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, hReq)
	fmt.Println("healthz:", w.Code)

	appReq := httptest.NewRequest(http.MethodGet, "/users", nil)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, appReq)
	fmt.Println("app:", w2.Code)

	// Output:
	// healthz: 200
	// app: 200
}
