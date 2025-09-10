package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonandereg/streamforge/internal/broker"
	"github.com/jonandereg/streamforge/internal/ingestor"
	sfmetrics "github.com/jonandereg/streamforge/internal/metrics"
	"github.com/jonandereg/streamforge/internal/obs"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// envCfg, err := config.LoadConfig()
	metricsPort := 2112
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "config error: %v\n", err)
	// 	os.Exit(1)
	// }

	obsCfg := obs.Config{
		ServiceName:    "streamforge-ingestor",
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

	o, shutdown, err := obs.Init(ctx, obsCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	defer shutdown(context.Background())

	sfmetrics.Register(o.PromRegistry)
	sfmetrics.Prime()
	obs.MustRegister(o.PromRegistry, broker.BrokerConnectTotal, broker.BrokerCloseTotal)

	mux := http.NewServeMux()
	m := o.HTTPMetrics
	mux.Handle(obsCfg.MetricsPath, o.MetricsHandler)
	mux.Handle(obsCfg.HealthPath, m.Wrap(obsCfg.HealthPath, o.HealthHandler))
	mux.Handle(obsCfg.ReadyPath, m.Wrap(obsCfg.ReadyPath, o.ReadyHandler.Handler()))
	obs.RegisterPprof(mux)

	addr := fmt.Sprintf(":%d", metricsPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	o.Logger.Info("starting service",
		zap.String("service", obsCfg.ServiceName),
		zap.String("version", obsCfg.ServiceVersion),
		zap.String("env", obsCfg.Env),
		zap.String("otlp_endpoint", obsCfg.OTLPEndpoint),
		zap.String("metrics_path", obsCfg.MetricsPath),
		zap.String("health_path", obsCfg.HealthPath),
		zap.String("ready_path", obsCfg.ReadyPath),
	)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	if err := ingestor.Run(ctx, o); err != nil {
		o.Logger.Fatal("ingestor start failed", zap.Error(err))
	}
	select {
	case <-ctx.Done():
		o.Logger.Info("shutdown signal recevied")
	case err := <-errCh:
		o.Logger.Error("metrics server error", zap.Error(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		o.Logger.Error("server shutdown error", zap.Error(err))
	} else {
		o.Logger.Info("server stopped cleanly")
	}

}
