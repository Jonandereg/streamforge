package model

import (
	"errors"
	"time"
)

// Tick is the normalized market event used across StreamForge.
// Fields are intentionally minimal and provider-agnostic.
type Tick struct {
	Symbol   string    `json:"symbol"`   // e.g., "AAPL", "EURUSD"
	Ts       time.Time `json:"ts"`       // event timestamp (UTC)
	Price    float64   `json:"price"`    // last trade/quote price
	Size     float64   `json:"size"`     // trade size (0 if unknown)
	Exchange string    `json:"exchange"` // source exchange/venue code ("" if unknown)
	SrcID    string    `json:"src_id"`   // provider/source identifier, e.g. "finnhub"
}

var (
	// ErrEmptySymbol is returned when a tick has an empty symbol.
	ErrEmptySymbol = errors.New("Tick: empty symbol")
	ErrBadTS       = errors.New("tick: zero timestamp")
	ErrBadPrice    = errors.New("tick: negative price")
	ErrBadSize     = errors.New("tick: negative size")
)

// Validate checks if the tick has valid field values.
func (t Tick) Validate() error {
	if t.Symbol == "" {
		return ErrEmptySymbol
	}
	if t.Ts.IsZero() {
		return ErrBadTS
	}
	if t.Price < 0 {
		return ErrBadPrice
	}
	if t.Size < 0 {
		return ErrBadSize
	}
	return nil
}
