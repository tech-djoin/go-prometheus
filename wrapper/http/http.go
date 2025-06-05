package http

import (
	"net/http"
	"strings"
	"time"

	goprometheus "github.com/tech-djoin/go-prometheus"
)

// MetricCollector is a middleware function for collecting metrics.
//
// This function returns a middleware function that wraps the next handler function
// framework's middleware chain. It measures the time taken to process the request and records
// relevant metrics including the HTTP request and latency.
func MetricCollector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture response status
		rw := &responseWriter{w, http.StatusOK}

		// Serve the next handler
		next.ServeHTTP(rw, r)

		cleanPath := strings.Split(r.URL.Path, "?")[0]

		// initialize params struct
		prometheus := goprometheus.Prometheus{
			Method:    r.Method,
			Path:      cleanPath,
			Code:      rw.status,
			StartTime: start,
		}

		// Record metrics
		prometheus.RecordMetric()
	})
}

// Custom responseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
