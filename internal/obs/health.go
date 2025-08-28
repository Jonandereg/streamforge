package obs

import (
	"net/http"
	"sync"
)

// Readiness tracks the readiness state of the application.
type Readiness struct {
	mu    sync.RWMutex
	ready bool
}

// NewHealthHandlers creates liveness and readiness handlers for health checks.
func NewHealthHandlers() (http.Handler, *Readiness) {
	live := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	Readiness := NewReadiness()
	return live, Readiness
}

// NewReadiness creates a new Readiness instance.
func NewReadiness() *Readiness {
	return &Readiness{
		ready: false,
	}
}

// SetReady marks the application as ready.
func (r *Readiness) SetReady() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ready = true
}

// SetNotReady marks the application as not ready.
func (r *Readiness) SetNotReady() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ready = false
}

// Handler returns an HTTP handler for readiness checks.
func (r *Readiness) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		if r.ready {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready"))
		}
	})
}
