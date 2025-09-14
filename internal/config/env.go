package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	DataProvider DataProvider
	Kafka        Kafka
	Processor    Processor
}

type Processor struct {
	NumWorkers    int
	QueueCapacity int
}

// DataProvider holds configuration values loaded from environment variables.
type DataProvider struct {
	Token   string
	BaseURL string
	WsURL   string
}

type Kafka struct {
	Brokers    []string
	GroupID    string
	TicksTopic string

	MinBytes int
	MaxBytes int
	MaxWait  time.Duration
}

// LoadConfig loads configuration values from environment variables.
func LoadConfig() (AppConfig, error) {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("warning: no .env file found")
	}
	dp := DataProvider{
		Token:   mustEnv("FINNHUB_TOKEN"),
		BaseURL: mustEnv("FINNHUB_BASE_URL"),
		WsURL:   mustEnv("FINNHUB_WS_URL"),
	}

	k := Kafka{
		Brokers:    splitAndTrim(mustEnv("KAFKA_BROKERS")),
		GroupID:    mustEnv("KAFKA_GROUP_ID"),
		TicksTopic: mustEnv("KAFKA_TICKS_TOPIC"),
		MinBytes:   mustEnvInt("KAFKA_MIN_BYTES"),
		MaxBytes:   mustEnvInt("KAFKA_MAX_BYTES"),
		MaxWait:    time.Duration(mustEnvInt("KAFKA_MAX_WAIT_MS")) * time.Millisecond,
	}

	p := Processor{
		NumWorkers:    mustEnvInt("TICKS_NUM_WORKERS"),
		QueueCapacity: mustEnvInt("TICKS_QUEUE_CAPACITY"),
	}

	return AppConfig{
		DataProvider: dp,
		Kafka:        k,
		Processor:    p,
	}, nil

}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Errorf("missing %s", key))
	}
	return v
}

func mustEnvInt(key string) int {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Errorf("missing %s", key))
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Errorf("invalid int for %s: %v", key, err))
	}
	return n
}
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
