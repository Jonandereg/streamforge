package events

import (
	"time"

	"github.com/jonandereg/streamforge/internal/model"
)

type KafkaMeta struct {
	Partition int
	Offset    int64
	Key       []byte
	Time      time.Time
}

type TickMsg struct {
	Tick  model.Tick
	Kafka KafkaMeta
}
