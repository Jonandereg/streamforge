package metrics

import (
	"github.com/jonandereg/streamforge/internal/obs"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	IngestorFetchTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_fetch_total",
			Help: "Total number of fetch attempts by the provider client.",
		},
	)
	IngestorFetchErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ingestor_fetch_errors_total",
			Help: "Total number of fetch errors, labeled by reason.",
		},
		[]string{"reason"},
	)

	// Total successful publishes to the broker
	IngestorPublishTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_publish_total",
			Help: "Total number of successfully published ticks.",
		},
	)

	// Publish errors partitioned by reason (e.g., retriable, non_retriable, timeout)
	IngestorPublishErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ingestor_publish_errors_total",
			Help: "Total number of publish errors, labeled by reason.",
		},
		[]string{"reason"},
	)

	// Latency around the publish call
	IngestorPublishLatencySeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "ingestor_publish_latency_seconds",
			Help: "Histogram of publish latency to the broker in seconds.",
			// Tweak later if needed; start with default-ish buckets
			Buckets: prometheus.DefBuckets,
		},
	)
)

func Register(reg *prometheus.Registry) {

	obs.MustRegister(reg,
		IngestorFetchTotal,
		IngestorFetchErrorsTotal,
		IngestorPublishTotal,
		IngestorPublishErrorsTotal,
		IngestorPublishLatencySeconds,
	)
}
