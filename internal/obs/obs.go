package obs

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// Obs holds observability components including logger, metrics, and tracing.
type Obs struct {
	Logger         *zap.Logger
	PromRegistry   *prometheus.Registry
	TracerProvider *sdktrace.TracerProvider

	// Optional helpers to mount endpoints
	MetricsHandler http.Handler
	HealthHandler  http.Handler
	ReadyHandler   http.Handler
}
