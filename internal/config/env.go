package config

import (
	"fmt"
	"os"
)

// Config holds configuration values loaded from environment variables.
type Config struct {
	DataProviderToken   string
	DataProviderBaseURL string
}

// LoadConfig loads configuration values from environment variables.
func LoadConfig() (Config, error) {
	token := os.Getenv("FINNHUB_TOKEN")

	if token == "" {
		return Config{}, fmt.Errorf(" missing FINNHUB_TOKEN")
	}
	baseURL := os.Getenv("FINNHUB_BASE_URL")
	if baseURL == "" {
		return Config{}, fmt.Errorf("missing FINNHUB_BASE_URL")
	}

	return Config{
		DataProviderToken:   token,
		DataProviderBaseURL: baseURL,
	}, nil

}
