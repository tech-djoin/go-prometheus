package goprometheus

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	appHttpRequest = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_http_request_totals",
		Help: "The total number of application request http",
	}, []string{"method", "path", "code"})

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

// MetricCollector is a middleware function for collecting metrics in an Echo framework application.
//
// This function returns a middleware function that wraps the next handler function in the Echo
// framework's middleware chain. It measures the time taken to process the request and records
// relevant metrics including the HTTP request and latency.
func MetricCollector() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			request := next(c)

			// Record metrics
			RecordHttpRequest(c.Request().Method, c.Path(), c.Response().Status)
			RecordLatency(c.Request().Method, c.Path(), start)

			return request
		}
	}
}
