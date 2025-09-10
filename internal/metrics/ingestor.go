// Package metrics defines Prometheus metrics for the StreamForge ingestor service.
package metrics

import (
	"github.com/jonandereg/streamforge/internal/obs"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// IngestorFetchTotal counts the total number of fetch attempts by the provider client.
	IngestorFetchTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_fetch_total",
			Help: "Total number of fetch attempts by the provider client.",
		},
	)
	// IngestorFetchErrorsTotal counts fetch errors partitioned by reason.
	IngestorFetchErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ingestor_fetch_errors_total",
			Help: "Total number of fetch errors, labeled by reason.",
		},
		[]string{"reason"},
	)

	// IngestorPublishTotal counts the total successful publishes to the broker.
	IngestorPublishTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_publish_total",
			Help: "Total number of successfully published ticks.",
		},
	)

	// IngestorPublishErrorsTotal counts publish errors partitioned by reason (e.g., retriable, non_retriable, timeout).
	IngestorPublishErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ingestor_publish_errors_total",
			Help: "Total number of publish errors, labeled by reason.",
		},
		[]string{"reason"},
	)

	// IngestorPublishLatencySeconds measures latency around the publish call.
	IngestorPublishLatencySeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "ingestor_publish_latency_seconds",
			Help: "Histogram of publish latency to the broker in seconds.",
			// Tweak later if needed; start with default-ish buckets
			Buckets: prometheus.DefBuckets,
		},
	)
	// IngestorProviderConnectTotal tracks provider connect attempts by status.
	IngestorProviderConnectTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ingestor_provider_connect_total",
			Help: "Provider connect attempts by status.",
		},
		[]string{"status"},
	)
	// IngestorProviderReconnectTotal tracks provider reconnect attemps
	IngestorProviderReconnectTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_provider_reconnect_total",
			Help: "Number of provider reconnects.",
		},
	)

	// IngestorBackpressureTotal counts times the publisher queue was full and we had to block or drop.
	IngestorBackpressureTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingestor_backpressure_total",
			Help: "Times the publisher queue was full and we had to block or drop.",
		},
	)
)

// Register registers all ingestor metrics with the provided Prometheus registry.
func Register(reg *prometheus.Registry) {

	obs.MustRegister(reg,
		IngestorFetchTotal,
		IngestorFetchErrorsTotal,
		IngestorPublishTotal,
		IngestorPublishErrorsTotal,
		IngestorPublishLatencySeconds,
		IngestorProviderConnectTotal,
		IngestorProviderReconnectTotal,
		IngestorBackpressureTotal,
	)
}

// Prime initializes all metric label combinations to ensure they appear in /metrics output.
func Prime() {
	for _, reason := range []string{"validation", "rate_limit", "http", "parse", "ws"} {
		IngestorFetchErrorsTotal.WithLabelValues(reason).Add(0)
	}
	for _, reason := range []string{"retriable", "non_retriable", "timeout", "marshal", "error"} {
		IngestorPublishErrorsTotal.WithLabelValues(reason).Add(0)
	}
	IngestorProviderConnectTotal.WithLabelValues("success").Add(0)
	IngestorProviderConnectTotal.WithLabelValues("failure").Add(0)

}
