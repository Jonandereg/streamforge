package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jonandereg/streamforge/internal/config"
	"github.com/jonandereg/streamforge/internal/events"
	"github.com/jonandereg/streamforge/internal/model"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type TickConsumer struct {
	cfg    config.Kafka
	reader *kafka.Reader
	log    *zap.Logger
}

func NewTickConsumer(cfg config.Kafka, log *zap.Logger) (*TickConsumer, error) {
	if len(cfg.Brokers) == 0 || cfg.Brokers[0] == "" {
		return nil, errors.New("kafka brokers missing")
	}
	if cfg.GroupID == "" || cfg.TicksTopic == "" {
		return nil, errors.New("kafka group or topic missing")
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.TicksTopic,
		MinBytes:    cfg.MinBytes,
		MaxBytes:    cfg.MaxBytes,
		MaxWait:     cfg.MaxWait,
		StartOffset: kafka.FirstOffset,
	})

	return &TickConsumer{
		cfg:    cfg,
		reader: r,
		log:    log.Named("tick-consumer"),
	}, nil
}

func (c *TickConsumer) Run(ctx context.Context, out chan<- events.TickMsg) error {
	c.log.Info("starting",
		zap.Strings("brokers", c.cfg.Brokers),
		zap.String("group", c.cfg.GroupID),
		zap.String("topic", c.cfg.TicksTopic),
	)
	defer func() {
		if err := c.reader.Close(); err != nil {
			c.log.Warn("reader close error", zap.Error(err))
		}
	}()

	backoff := 200 * time.Millisecond

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				c.log.Info("context closed, exiting")
				return nil
			}
			c.log.Warn("fetch error", zap.Error(err))
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil
			}
			continue
		}
		var t model.Tick
		if err := json.Unmarshal(m.Value, &t); err != nil {
			c.log.Warn("json decode error,skipping",
				zap.Int("partition", m.Partition),
				zap.Int64("offset", m.Offset),
				zap.Error(err),
			)
			_ = c.reader.CommitMessages(ctx, m)
			continue

		}

		msg := events.TickMsg{
			Tick: t,
			Kafka: events.KafkaMeta{
				Partition: m.Partition,
				Offset:    m.Offset,
				Key:       m.Key,
				Time:      m.Time,
			},
		}
		select {
		case out <- msg:
			if err := c.reader.CommitMessages(ctx, m); err != nil {
				c.log.Warn("commit error", zap.Error(err))
			}
		case <-ctx.Done():
			return nil

		}
	}

}
