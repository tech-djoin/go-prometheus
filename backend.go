package goprometheus

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricsBackend is responsible for persisting metrics emitted by this package.
type MetricsBackend interface {
	RecordHttpRequest(method string, path string, code int)
	RecordLatency(method string, path string, elapsedSeconds float64)
}

type backendHolder struct {
	backend MetricsBackend
}

var (
	backend                  atomic.Value
	defaultMetricsRegistered atomic.Bool
)

func init() {
	defaultMetricsRegistered.Store(true)
	backend.Store(&backendHolder{backend: prometheusBackend{}})
}

func currentBackend() MetricsBackend {
	return backend.Load().(*backendHolder).backend
}

// UseBackend replaces the metrics backend used by this package. It is safe for
// concurrent use.
func UseBackend(b MetricsBackend) {
	if b == nil {
		panic("goprometheus: backend cannot be nil")
	}
	switch b.(type) {
	case prometheusBackend, *prometheusBackend:
		ensureDefaultMetricsRegistered()
	default:
		ensureDefaultMetricsUnregistered()
	}
	backend.Store(&backendHolder{backend: b})
}

// ResetBackend switches the metrics backend back to the built-in Prometheus
// implementation.
func ResetBackend() {
	ensureDefaultMetricsRegistered()
	backend.Store(&backendHolder{backend: prometheusBackend{}})
}

func ensureDefaultMetricsRegistered() {
	if defaultMetricsRegistered.Load() {
		return
	}
	prometheus.MustRegister(appHttpRequest, appHttpLatency)
	defaultMetricsRegistered.Store(true)
}

func ensureDefaultMetricsUnregistered() {
	if !defaultMetricsRegistered.Load() {
		return
	}
	// Unregister both metrics, ignoring the return value so we always mark them as unregistered.
	prometheus.Unregister(appHttpRequest)
	prometheus.Unregister(appHttpLatency)
	defaultMetricsRegistered.Store(false)
}
