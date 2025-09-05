package obs

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HTTPMetrics provides Prometheus metrics for HTTP requests including
// request counts and duration histograms.
type HTTPMetrics struct {
	requests  *prometheus.CounterVec
	durations *prometheus.HistogramVec
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// NewHTTPMetrics creates a new HTTPMetrics instance with request counter
// and duration histogram metrics registered to the provided registry.
func NewHTTPMetrics(reg *prometheus.Registry) *HTTPMetrics {
	m := &HTTPMetrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "streamforge_http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "route", "status"},
		),
		durations: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "streamforge_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
			},
			[]string{"method", "route", "status"},
		),
	}

	MustRegister(reg, m.requests, m.durations)

	return m
}

// Wrap returns an HTTP handler that wraps the provided handler with metrics
// collection, recording request counts and durations for the specified route.
func (m *HTTPMetrics) Wrap(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(ww, r)
		method := r.Method
		status := strconv.Itoa(ww.status)

		m.requests.WithLabelValues(method, route, status).Inc()
		m.durations.WithLabelValues(method, route, status).Observe(time.Since(start).Seconds())
	})
}
