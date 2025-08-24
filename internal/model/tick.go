package model

import "time"

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
