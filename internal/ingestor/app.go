package ingestor

import (
	"context"
	"time"

	"github.com/jonandereg/streamforge/internal/broker"
	"github.com/jonandereg/streamforge/internal/config"
	sfmetrics "github.com/jonandereg/streamforge/internal/metrics"
	"github.com/jonandereg/streamforge/internal/obs"
	"github.com/jonandereg/streamforge/internal/providers/finnhub"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func Run(ctx context.Context, o *obs.Obs) error {

	bcfg := broker.Config{
		Brokers:       []string{"localhost:29092"},
		Topic:         "ticks",
		ClientID:      "streamforge-ingestor",
		Acks:          -1, // all
		BatchTimeout:  5 * time.Millisecond,
		BatchBytes:    1_048_576,
		Compression:   kafka.Lz4.Codec(),
		RetryAttempts: 5,
		RetryBackoff:  100 * time.Millisecond,
	}

	prod, err := broker.NewProducer(ctx, bcfg)
	if err != nil {
		o.Logger.Error("failed to connect to broker", zap.Error(err))
		return err
	}
	defer prod.Close()
	o.ReadyHandler.SetReady()
	o.Logger.Info("broker connected; readiness set")

	envConfig, _ := config.LoadConfig()

	symbols := []string{
		"APPL",
		"MSFT",
		"BINANCE:BTCUSDT",
	}
	provCfg := finnhub.WSConfig{
		BaseURL:       envConfig.DataProviderWsURL,
		APIKey:        envConfig.DataProviderToken,
		Symbols:       symbols,
		ReconnectBase: 200 * time.Millisecond,
		ReconnectMax:  5 * time.Second,
	}

	prov := finnhub.New(provCfg, o.Logger)

	ticksCh, errsCh := prov.Start(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-ticksCh:
				if !ok {
					return
				}
				sfmetrics.IngestorFetchTotal.Inc()
				if err := prod.Publish(ctx, t); err != nil {
					o.Logger.Error("publish failed",
						zap.String("symbol", t.Symbol),
						zap.Time("ts", t.Ts),
						zap.Error(err),
					)
					continue
				}
				o.Logger.Debug("published tick",
					zap.String("symbol", t.Symbol),
					zap.Time("ts", t.Ts),
					zap.Float64("price", t.Price),
				)
			case err, ok := <-errsCh:
				if !ok {
					return
				}
				sfmetrics.IngestorFetchErrorsTotal.WithLabelValues("ws").Inc()
				o.Logger.Warn("provider error", zap.Error(err))
			}
		}
	}()

	<-ctx.Done()

	o.Logger.Info("shutdown signal received, closing producer")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- prod.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			o.Logger.Error("failed to close producer", zap.Error(err))
		} else {
			o.Logger.Info("producer closed cleanly")
		}
	case <-shutdownCtx.Done():

		o.Logger.Warn("producer close timed out; exiting")
	}

	return nil
}
