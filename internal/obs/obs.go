package obs

import (
	"context"
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
	MetricsHandler http.Handler
	HealthHandler  http.Handler
	ReadyHandler   *Readiness
	HTTPMetrics    *HTTPMetrics
}

// Init sets up observability components (logger, tracing, metrics, health, readiness).
// It returns an Obs struct with handlers and clients, plus a shutdown function
// that should be called on service exit to flush logs and traces.
func Init(ctx context.Context, cfg Config) (*Obs, func(context.Context) error, error) {
	lg, err := NewLogger(cfg)
	if err != nil {
		return nil, nil, err
	}

	tp, err := NewTracerProvider(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	reg, metricsH, err := NewMetricsRegistry(cfg)
	if err != nil {
		return nil, nil, err
	}
	healthH, readyH := NewHealthHandlers()

	httpM := NewHTTPMetrics(reg)

	o := &Obs{
		Logger:         lg,
		TracerProvider: tp,
		PromRegistry:   reg,
		MetricsHandler: metricsH,
		HealthHandler:  healthH,
		ReadyHandler:   readyH,
		HTTPMetrics:    httpM,
	}

	shutdown := func(ctx context.Context) error {
		var retErr error
		_ = o.Logger.Sync()
		if o.TracerProvider != nil {
			if err := o.TracerProvider.Shutdown(ctx); err != nil {
				retErr = err
			}
		}
		return retErr
	}

	return o, shutdown, nil
}
