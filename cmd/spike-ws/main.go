//go:build spike

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jonandereg/streamforge/internal/config"
	"github.com/jonandereg/streamforge/internal/model"
)

type envelope struct {
	Type string       `json:"type"`
	Data []tradeEvent `json:"data"`
}

type tradeEvent struct {
	Price    float64 `json:"p"`
	Symbol   string  `json:"s"`
	TSMS     int64   `json:"t"`           // epoch ms
	Size     float64 `json:"v"`           // trade size/volume
	Exchange string  `json:"x,omitempty"` // not always present
	// conditions "c" omitted for the spike
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// CLI: -symbols AAPL,MSFT
	var symbolsCSV string
	flag.StringVar(&symbolsCSV, "symbols", "AAPL,MSFT", "Comma-separated symbols")
	flag.Parse()
	symbols := splitAndTrim(symbolsCSV)
	if len(symbols) == 0 {
		log.Error("no symbols provided")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("config error", "err", err)
		os.Exit(1)
	}

	// Build ws URL: e.g. wss://ws.finnhub.io/?token=XYZ
	u, err := url.Parse(cfg.DataProviderBaseURL)
	if err != nil {
		log.Error("bad FINNHUB_BASE_URL", "url", cfg.DataProviderBaseURL, "err", err)
		os.Exit(1)
	}
	q := u.Query()
	q.Set("token", cfg.DataProviderToken)
	u.RawQuery = q.Encode()

	log.Info("connecting", "url", u.Redacted())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl+C / SIGTERM
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(ch)
		<-ch
		cancel()
	}()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("dial failed", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Subscribe
	for _, s := range symbols {
		msg := fmt.Sprintf(`{"type":"subscribe","symbol":"%s"}`, s)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Error("subscribe failed", "symbol", s, "err", err)
			os.Exit(1)
		}
		log.Info("subscribed", "symbol", s)
	}

	// Reader loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Error("read failed", "err", err)
			return
		}

		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			log.Warn("unmarshal", "err", err)
			continue
		}
		if env.Type != "trade" {
			continue // ignore ping/info
		}

		for _, te := range env.Data {
			t := model.Tick{
				Symbol:   te.Symbol,
				Ts:       time.UnixMilli(te.TSMS).UTC(),
				Price:    te.Price,
				Size:     te.Size,
				Exchange: te.Exchange, // may be ""
				SrcID:    "finnhub",
			}
			t = model.NormalizeTick(t)

			// print normalized JSON
			out, _ := json.Marshal(t)
			fmt.Println(string(out))
		}
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
