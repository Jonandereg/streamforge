package broker

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

var (
	BrokerConnectTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ingestor_broker_connect_total",
			Help: "Total broker connection attempts by status.",
		},
		[]string{"status"},
	)

	BrokerCloseTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ingestor_broker_close_total",
			Help: "Total broker close attempts by status.",
		},
		[]string{"status"},
	)
)

type Config struct {
	Brokers       []string
	Topic         string
	ClientID      string
	Acks          kafka.RequiredAcks
	BatchTimeout  time.Duration
	BatchBytes    int
	Compression   kafka.CompressionCodec
	RetryAttempts int
	RetryBackoff  time.Duration
}

// NewProducer creates a new Kafka producer and pings the broker.
func NewProducer(ctx context.Context, cfg Config) (*Producer, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{

		Brokers:          cfg.Brokers,
		Topic:            cfg.Topic,
		Balancer:         &kafka.Hash{},
		RequiredAcks:     int(cfg.Acks),
		Async:            false,
		BatchTimeout:     cfg.BatchTimeout,
		BatchBytes:       cfg.BatchBytes,
		Logger:           nil,
		ErrorLogger:      nil,
		CompressionCodec: cfg.Compression,
	})

	conn, err := kafka.DialContext(ctx, "tcp", cfg.Brokers[0])
	if err != nil {
		BrokerConnectTotal.WithLabelValues("failure").Inc()
		writer.Close()
		return nil, err
	}
	_ = conn.Close()
	BrokerConnectTotal.WithLabelValues("success").Inc()

	return &Producer{writer: writer}, nil
}

// Close flushes and closes the producer.
func (p *Producer) Close() error {
	err := p.writer.Close()
	if err != nil {
		BrokerCloseTotal.WithLabelValues("failure").Inc()
		return err
	}
	BrokerCloseTotal.WithLabelValues("success").Inc()
	return nil
}
