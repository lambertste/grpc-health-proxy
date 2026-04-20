// Command grpc-health-proxy is a lightweight sidecar that translates gRPC
// health checks to HTTP endpoints for legacy load balancers.
//
// It polls a gRPC service's health endpoint and exposes the result over HTTP
// so that infrastructure that cannot speak gRPC (e.g. classic AWS ELBs,
// HAProxy, or Kubernetes httpGet probes) can still participate in health-based
// routing.
//
// Configuration is driven entirely by environment variables; see
// config/config.go for the full list.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/grconfig"
	"github"
	"github.com/your-org/grpc-health-proxy/healthz"
	"github.com/your-org/grpc-health-proxy/metrics"
	"github.com/your-org/grpc-health-proxy/middleware"
	"github.com/your-org/grpc-health-proxy/server"
	"github.com/your-org/grpc-health-proxy/tls"
	"github.com/your-org/grpc-health-proxy/tracing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// --- TLS (optional) ---------------------------------------------------
	tlsCfg, err := tls.ConfigFromEnv()
	if err != nil {
		log.Fatalf("tls: %v", err)
	}

	// --- gRPC health checker ----------------------------------------------
	checker, err := health.NewChecker(cfg)
	if err != nil {
		log.Fatalf("checker: %v", err)
	}
	defer checker.Close()

	// --- Prometheus metrics -----------------------------------------------
	m := metrics.New()

	// --- Drain / graceful-shutdown barrier --------------------------------
	drainer := drain.New(cfg.ShutdownDeadline)

	// --- Liveness / readiness self-probe ----------------------------------
	hz := healthz.New()

	// --- HTTP server ------------------------------------------------------
	srv := server.New(cfg, checker)

	mux := http.NewServeMux()
	mux.Handle("/healthz/", http.StripPrefix("/healthz", hz.Handler()))
	mux.Handle("/metrics", m.Handler())
	mux.Handle("/", srv.Handler())

	chain := middleware.Chain(
		tracing.Middleware,
		drainer.Middleware,
		middleware.Logging(log.New(os.Stdout, "", log.LstdFlags)),
		middleware.Metrics(m),
	)

	httpServer := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      chain(mux),
		TLSConfig:    tlsCfg,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// --- Start ------------------------------------------------------------
	go func() {
		log.Printf("listening on %s", cfg.ListenAddr)
		var serveErr error
		if tlsCfg != nil {
			serveErr = httpServer.ListenAndServeTLS("", "")
		} else {
			serveErr = httpServer.ListenAndServe()
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Fatalf("http server: %v", serveErr)
		}
	}()

	hz.MarkReady()

	// --- Graceful shutdown ------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutdown signal received; draining…")
	drainer.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownDeadline+5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
	log.Println("stopped")
}
