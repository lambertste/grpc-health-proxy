package reflect_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/salrashid123/grpc-health-proxy/reflect"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNew_PanicsOnNilNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, got none")
		}
	}()
	reflect.New(nil)
}

func TestServeHTTP_SetsDefaultHeaders(t *testing.T) {
	rfl := reflect.New(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
	req.Proto = "HTTP/1.1"
	req.RemoteAddr = "127.0.0.1:9000"
	rec := httptest.NewRecorder()

	rfl.ServeHTTP(rec, req)

	assertHeader(t, rec, "X-Reflect-Method", "GET")
	assertHeader(t, rec, "X-Reflect-Path", "/foo/bar")
	assertHeader(t, rec, "X-Reflect-Remote", "127.0.0.1:9000")
	assertHeader(t, rec, "X-Reflect-Proto", "HTTP/1.1")
}

func TestServeHTTP_DelegatesToNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})
	rfl := reflect.New(next)

	rec := httptest.NewRecorder()
	rfl.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
}

func TestServeHTTP_CustomMethodHeader(t *testing.T) {
	rfl := reflect.New(
		http.HandlerFunc(okHandler),
		reflect.WithMethodHeader("X-My-Method"),
	)
	rec := httptest.NewRecorder()
	rfl.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/", nil))

	assertHeader(t, rec, "X-My-Method", "DELETE")
	if rec.Header().Get("X-Reflect-Method") != "" {
		t.Fatal("default method header should not be set when overridden")
	}
}

func TestServeHTTP_CustomPathHeader(t *testing.T) {
	rfl := reflect.New(
		http.HandlerFunc(okHandler),
		reflect.WithPathHeader("X-My-Path"),
	)
	rec := httptest.NewRecorder()
	rfl.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/custom", nil))

	assertHeader(t, rec, "X-My-Path", "/custom")
}

func TestServeHTTP_CustomRemoteHeader(t *testing.T) {
	rfl := reflect.New(
		http.HandlerFunc(okHandler),
		reflect.WithRemoteHeader("X-My-Remote"),
	)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	rfl.ServeHTTP(rec, req)

	assertHeader(t, rec, "X-My-Remote", "10.0.0.1:1234")
}

func TestServeHTTP_CustomProtoHeader(t *testing.T) {
	rfl := reflect.New(
		http.HandlerFunc(okHandler),
		reflect.WithProtoHeader("X-My-Proto"),
	)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Proto = "HTTP/2.0"
	rec := httptest.NewRecorder()
	rfl.ServeHTTP(rec, req)

	assertHeader(t, rec, "X-My-Proto", "HTTP/2.0")
}

func assertHeader(t *testing.T, rec *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	got := rec.Header().Get(key)
	if got != want {
		t.Fatalf("header %q: want %q, got %q", key, want, got)
	}
}
