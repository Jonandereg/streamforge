package ingestor

import (
	"context"
	"time"

	"github.com/jonandereg/streamforge/internal/broker"
	"github.com/jonandereg/streamforge/internal/obs"
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
