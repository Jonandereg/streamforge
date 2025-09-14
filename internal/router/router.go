package router

import (
	"context"
	"hash/fnv"

	"github.com/jonandereg/streamforge/internal/events"
)

// StartRouter creates numWorkers output channels and starts a goroutine that
// routes each TickMsg to a deterministic worker index based on Tick.Symbol.

func StartRouter(ctx context.Context, in <-chan events.TickMsg, numWorkers, queueCap int, onDrop func(events.TickMsg)) []chan events.TickMsg {
	if numWorkers <= 0 {
		numWorkers = 1
	}
	if queueCap <= 0 {
		queueCap = 0
	}

	outs := make([]chan events.TickMsg, numWorkers)
	for i := range outs {
		outs[i] = make(chan events.TickMsg, queueCap)
	}

	go func() {
		defer func() {
			for _, ch := range outs {
				close(ch)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-in:
				if !ok {
					return
				}
				idx := workerIndex(msg.Tick.Symbol, numWorkers)
				select {
				case outs[idx] <- msg:
				default:
					if onDrop != nil {
						onDrop(msg)
					}
				}

			}
		}
	}()

	return outs
}

// workerIndex maps a symbol to a stable worker index in [0, numWorkers).
func workerIndex(symbol string, numWorkers int) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(symbol))
	return int(h.Sum32() % uint32(numWorkers))

}
