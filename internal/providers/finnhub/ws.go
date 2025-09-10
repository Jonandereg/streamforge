// Package finnhub provides a WebSocket client for the Finnhub market data API.
package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jonandereg/streamforge/internal/model"
	"go.uber.org/zap"
)

// WSConfig holds configuration for the Finnhub WebSocket client.
type WSConfig struct {
	BaseURL       string
	APIKey        string
	Symbols       []string
	ReconnectBase time.Duration // e.g. 200 * time.Millisecond
	ReconnectMax  time.Duration // e.g. 5 * time.Second
}

// Provider implements a Finnhub WebSocket client for streaming market data.
type Provider struct {
	cfg WSConfig
	log *zap.Logger
}

// New creates a new Finnhub WebSocket provider with the given configuration.
func New(cfg WSConfig, log *zap.Logger) *Provider {
	syms := make([]string, 0, len(cfg.Symbols))

	for _, s := range cfg.Symbols {
		s = strings.ToUpper(strings.TrimSpace(s))
		if s != "" {
			syms = append(syms, s)
		}

	}
	cfg.Symbols = syms

	if cfg.ReconnectBase <= 0 {
		cfg.ReconnectBase = 200 * time.Millisecond
	}
	if cfg.ReconnectMax <= 0 {
		cfg.ReconnectMax = 5 * time.Second
	}
	return &Provider{cfg: cfg, log: log}
}

type envelope struct {
	Type string       `json:"type"`
	Data []tradeEvent `json:"data"`
}

type tradeEvent struct {
	Price    float64 `json:"p"`
	Symbol   string  `json:"s"`
	TSMS     int64   `json:"t"`           // epoch ms
	Size     float64 `json:"v"`           // trade size/volume
	Exchange string  `json:"x,omitempty"` // optional
}

// Start begins streaming market data from Finnhub WebSocket API.
func (p *Provider) Start(ctx context.Context) (<-chan model.Tick, <-chan error) {
	ticks := make(chan model.Tick, 1024)
	errs := make(chan error, 16)

	go func() {
		defer close(ticks)
		defer close(errs)
		backoff := p.cfg.ReconnectBase
		for {
			u, err := url.Parse(p.cfg.BaseURL)
			if err != nil {
				errs <- fmt.Errorf("finnhub: bad base url: %w", err)
				return
			}
			q := u.Query()
			q.Set("token", p.cfg.APIKey)
			u.RawQuery = q.Encode()
			p.log.Info("finnhub: connecting", zap.String("url", u.Redacted()))
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				p.log.Warn("finnhub: dial failed, will retry",
					zap.Error(err),
					zap.Duration("sleep", backoff),
				)

				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return
				}
				if backoff < p.cfg.ReconnectMax {
					backoff *= 2
					if backoff > p.cfg.ReconnectMax {
						backoff = p.cfg.ReconnectMax
					}
				}
				continue
			}
			backoff = p.cfg.ReconnectBase

			//subscribe to symbols
			for _, s := range p.cfg.Symbols {
				msg := fmt.Sprintf(`{"type":"subscribe","symbol":"%s"}`, s)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					p.log.Warn("finnhub: subscribe failed", zap.String("symbol", s), zap.Error(err))
				} else {
					p.log.Info("finnhub: subscribed", zap.String("symbol", s))
				}
			}

			// Read loop blocks here until ctx cancel or read error
			readErr := p.readLoop(ctx, conn, ticks, errs)

			_ = conn.Close()
			if ctx.Err() != nil {
				return
			}

			p.log.Warn("finnhub: connection closed, will reconnect",
				zap.Error(readErr),
				zap.Duration("sleep", backoff),
			)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			}
			if backoff < p.cfg.ReconnectMax {
				backoff *= 2
				if backoff > p.cfg.ReconnectMax {
					backoff = p.cfg.ReconnectMax
				}
			}

		}

	}()

	return ticks, errs
}

func (p *Provider) readLoop(ctx context.Context, conn *websocket.Conn, ticks chan<- model.Tick, errs chan<- error) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			errs <- fmt.Errorf("finnhub: unmarshal: %w", err)
			continue
		}
		if env.Type != "trade" {
			continue // ignore ping/info
		}
		for _, te := range env.Data {
			t := model.Tick{
				Symbol:   strings.ToUpper(strings.TrimSpace(te.Symbol)),
				Ts:       time.UnixMilli(te.TSMS).UTC(),
				Price:    te.Price,
				Size:     te.Size,
				Exchange: te.Exchange,
				SrcID:    "finnhub",
			}
			t = model.NormalizeTick(t)
			select {
			case ticks <- t:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

}
