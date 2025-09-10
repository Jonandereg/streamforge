package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds configuration values loaded from environment variables.
type Config struct {
	DataProviderToken   string
	DataProviderBaseURL string
	DataProviderWsURL   string
}

// LoadConfig loads configuration values from environment variables.
func LoadConfig() (Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("warning: no .env file found")
	}
	token := os.Getenv("FINNHUB_TOKEN")

	if token == "" {
		return Config{}, fmt.Errorf(" missing FINNHUB_TOKEN")
	}
	baseURL := os.Getenv("FINNHUB_BASE_URL")
	if baseURL == "" {
		return Config{}, fmt.Errorf("missing FINNHUB_BASE_URL")
	}
	wsURL := os.Getenv("FINNHUB_WS_URL")

	if wsURL == "" {
		return Config{}, fmt.Errorf("missing FINNHUB_WS_URL")
	}

	return Config{
		DataProviderToken:   token,
		DataProviderBaseURL: baseURL,
		DataProviderWsURL:   wsURL,
	}, nil

}
