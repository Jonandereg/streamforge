package obs

// Config holds observability configuration settings.
type Config struct {
	ServiceName    string // e.g., "streamforge-ingestor"
	ServiceVersion string // from pkg/version
	Env            string // dev|staging|prod
	LogLevel       string // debug|info|warn|error
	LogJSON        bool   // true in prod
	OTLPEndpoint   string // Jaeger OTLP HTTP, e.g., http://jaeger:4318
	EnablePprof    bool
	MetricsPath    string // default "/metrics"
	HealthPath     string // default "/healthz"
	ReadyPath      string // default "/readyz"
}
