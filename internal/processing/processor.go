package processing

import (
	"context"

	"github.com/jonandereg/streamforge/internal/events"
	"go.uber.org/zap"
)

type NoopProcessor struct {
	Log *zap.Logger
}

func (p *NoopProcessor) Process(ctx context.Context, msg events.TickMsg) error {

	if msg.Tick.Symbol == "" {
		if p.Log != nil {
			p.Log.Warn("drop: empty symbol",
				zap.Int("partition", msg.Kafka.Partition),
				zap.Int64("offset", msg.Kafka.Offset),
			)
		}
		return nil
	}

	if p.Log != nil {
		p.Log.Debug("processed tick",
			zap.String("symbol", msg.Tick.Symbol),
			zap.Float64("price", msg.Tick.Price),
			zap.Int("partition", msg.Kafka.Partition),
			zap.Int64("offset", msg.Kafka.Offset),
		)
	}
	return nil
}
