package providers

import (
	"context"

	"github.com/jonandereg/streamforge/internal/model"
)

type Provider interface {
	// Start begins streaming ticks until ctx is cancelled.
	// Returns two read-only channels: ticks and errors.
	Start(ctx context.Context) (<-chan model.Tick, <-chan error)
}
