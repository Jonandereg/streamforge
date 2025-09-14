package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonandereg/streamforge/internal/config"
	"github.com/jonandereg/streamforge/internal/consumer"
	"github.com/jonandereg/streamforge/internal/events"
	sfmetrics "github.com/jonandereg/streamforge/internal/metrics"
	"github.com/jonandereg/streamforge/internal/obs"
	"github.com/jonandereg/streamforge/internal/processing"
	"github.com/jonandereg/streamforge/internal/router"
	"github.com/jonandereg/streamforge/internal/worker"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metricsPort := 2113

	obsCfg := obs.Config{
		ServiceName:    "streamforge-ticks-processor",
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
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "shutdown error: %v\n", err)
		}
	}()

	sfmetrics.Register(o.PromRegistry)
	sfmetrics.Prime()
	mux := http.NewServeMux()
	m := o.HTTPMetrics
	mux.Handle(obsCfg.MetricsPath, o.MetricsHandler)
	mux.Handle(obsCfg.HealthPath, m.Wrap(obsCfg.HealthPath, o.HealthHandler))
	mux.Handle(obsCfg.ReadyPath, m.Wrap(obsCfg.ReadyPath, o.ReadyHandler.Handler()))
	obs.RegisterPprof(mux)

	addr := fmt.Sprintf(":%d", metricsPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
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

	envCfg, _ := config.LoadConfig()

	ticksCh := make(chan events.TickMsg, 1024)
	cons, err := consumer.NewTickConsumer(envCfg.Kafka, o.Logger)
	if err != nil {
		o.Logger.Fatal("consumer init failed", zap.Error(err))
	}
	go func() {
		if err := cons.Run(ctx, ticksCh); err != nil {
			o.Logger.Error("consumer stopped with error", zap.Error(err))
		}
		close(ticksCh)
	}()
	onDrop := func(m events.TickMsg) {
		o.Logger.Warn("router drop: worker queue full", zap.String("symbol", m.Tick.Symbol))
	}

	outs := router.StartRouter(ctx, ticksCh, envCfg.Processor.NumWorkers, envCfg.Processor.QueueCapacity, onDrop)

	proc := &processing.NoopProcessor{
		Log: o.Logger,
	}

	worker.StartWorkers(ctx, outs, proc, o.Logger)

	// ---- SHUTDOWN ----
	select {
	case <-ctx.Done():
		o.Logger.Info("shutdown signal received")
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
