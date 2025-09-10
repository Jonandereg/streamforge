// Package providers defines interfaces for market data providers.
package providers

import (
	"context"

	"github.com/jonandereg/streamforge/internal/model"
)

// Provider defines the interface for market data providers.
type Provider interface {
	// Start begins streaming ticks until ctx is cancelled.
	// Returns two read-only channels: ticks and errors.
	Start(ctx context.Context) (<-chan model.Tick, <-chan error)
}
