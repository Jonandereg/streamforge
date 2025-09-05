//go:build obsdemo

package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/jonandereg/streamforge/internal/obs"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := obs.Config{
		ServiceName:    "streamforge-obs-demo",
		ServiceVersion: "0.1.0",
		Env:            "dev",
		LogLevel:       "debug",
		LogJSON:        false,
		OTLPEndpoint:   "localhost:4318",
		EnablePprof:    true,
		MetricsPath:    "/metrics",
		HealthPath:     "/healthz",
		ReadyPath:      "/readyz",
	}

	o, shutdown, err := obs.Init(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer shutdown(context.Background())

	mux := http.NewServeMux()
	m := o.HTTPMetrics

	mux.Handle(cfg.MetricsPath, o.MetricsHandler)
	mux.Handle(cfg.HealthPath, m.Wrap(cfg.HealthPath, o.HealthHandler))
	mux.Handle(cfg.ReadyPath, m.Wrap(cfg.ReadyPath, o.ReadyHandler.Handler()))

	o.ReadyHandler.SetReady()

	obs.RegisterPprof(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tracer := o.TracerProvider.Tracer("demo")
		tracerCtx, span := tracer.Start(r.Context(), "demo-request")
		defer span.End()

		log := obs.WithContext(tracerCtx, o.Logger)
		log.Info("Handling request")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello World!"))

	})

	srv := &http.Server{
		Addr: ":8080", Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			o.Logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	o.Logger.Info("shutting down")
	srv.Shutdown(context.Background())

}
