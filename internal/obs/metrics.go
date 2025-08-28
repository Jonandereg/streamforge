package obs

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewMetricsRegistry creates a new Prometheus registry with default collectors and handler.
func NewMetricsRegistry(_ Config) (*prometheus.Registry, http.Handler, error) {
	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics:   true,
		MaxRequestsInFlight: 5,
		Timeout:             10 * time.Second,
	})
	return reg, h, nil
}

// MustRegister registers multiple collectors with the registry, panicking on error.
func MustRegister(reg *prometheus.Registry, collectors ...prometheus.Collector) {
	for _, c := range collectors {
		reg.MustRegister(c)
	}
}
