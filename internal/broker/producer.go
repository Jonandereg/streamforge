package broker

import (
	"context"
	"encoding/json"
	"time"

	sfmetrics "github.com/jonandereg/streamforge/internal/metrics"
	"github.com/jonandereg/streamforge/internal/model"
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

// Publish sends one normalized Tick to Kafka with key=symbol and JSON value.
func (p *Producer) Publish(ctx context.Context, t model.Tick) error {
	start := time.Now()

	val, err := json.Marshal(t)
	if err != nil {
		sfmetrics.IngestorPublishErrorsTotal.WithLabelValues("marshal").Inc()
		return err
	}

	msg := kafka.Message{
		Key:   []byte(t.Symbol),
		Value: val,
		Headers: []kafka.Header{
			{Key: "content-type", Value: []byte("application/json")},
			{Key: "src_id", Value: []byte(t.SrcID)},
			{Key: "normalize_ver", Value: []byte("v1")},
		},
		Time: t.Ts,
	}

	err = p.writer.WriteMessages(ctx, msg)
	sfmetrics.IngestorPublishLatencySeconds.Observe(float64(time.Since(start).Seconds()))

	if err != nil {
		sfmetrics.IngestorPublishErrorsTotal.WithLabelValues("error").Inc()
		return err
	}

	sfmetrics.IngestorPublishTotal.Inc()
	return nil
}
