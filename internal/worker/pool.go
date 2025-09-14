package worker

import (
	"context"

	"github.com/jonandereg/streamforge/internal/events"
	"go.uber.org/zap"
)

// Processor abstracts business processing for a TickMsg.
type Processor interface {
	Process(ctx context.Context, msg events.TickMsg) error
}

func StartWorkers(ctx context.Context, inputs []chan events.TickMsg, proc Processor, log *zap.Logger) {
	for i := range inputs {
		i := i
		ch := inputs[i]
		go func() {
			wlog := log.Named("worker").With(zap.Int("id", i))
			wlog.Info("started")
			defer wlog.Info("stopped")
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-ch:
					if !ok {
						return
					}
					if err := proc.Process(ctx, msg); err != nil {
						// retry / policies can be added here later
						wlog.Warn("process failed",
							zap.String("symbol", msg.Tick.Symbol),
							zap.Error(err),
						)
					}
				}
			}
		}()
	}
}
