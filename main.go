package goprometheus

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus struct represents a params needed to using all metrics
type Prometheus struct {
	Method    string
	Path      string
	Code      int
	StartTime time.Time
}

// Prometheus Metric
var (
	// appHttpRequest is a counter metric to record total number of application request
	// with labels of method, path, and code
	appHttpRequest = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_http_request_totals",
		Help: "The total number of application request http",
	}, []string{"method", "path", "code"})

	// appHttpLatency is a histogram metric to record latency of application request
	// using buckets in range of 0.1, 0.3, 0.5, 0.7, and 0.9
	// with labels of method and path
	appHttpLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "app_http_request_latency_seconds",
			Help: "Latency of HTTP requests.",
			// Define the desired histogram buckets.
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 0.9},
		},
		[]string{"method", "path"},
	)
)

// RecordHttpRequest records an HTTP request.
//
// This function takes the HTTP method, path, and response code of an HTTP request
// as input parameters. It increments the appHttpRequest metric with the provided
// method, path, and code labels.
func RecordHttpRequest(method string, path string, code int) {
	appHttpRequest.WithLabelValues(method, path, strconv.Itoa(code)).Inc()
}

// RecordLatency records the latency of an HTTP request.
//
// This function takes the HTTP method, path, and the start time of the request
// as input parameters. It calculates the elapsed time since the start time
// and records the latency using the appHttpLatency metric.
func RecordLatency(method string, path string, start time.Time) {
	elapsed := time.Since(start).Seconds()
	appHttpLatency.WithLabelValues(method, path).Observe(elapsed)
}

// RecordMetric is a function that records all metric
//
// It calls the RecordHttpRequest function with the HTTP method, path, and response code,
// and also calls the RecordLatency function with the HTTP method, path, and start time.
func (p *Prometheus) RecordMetric() {
	RecordHttpRequest(p.Method, p.Path, p.Code)
	RecordLatency(p.Method, p.Path, p.StartTime)
}
