package echo

import (
	"time"

	"github.com/labstack/echo/v4"
	goprometheus "github.com/tech-djoin/go-prometheus"
)

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

			// initialize params struct
			prometheus := goprometheus.Prometheus{
				Method:    c.Request().Method,
				Path:      c.Path(),
				Code:      c.Response().Status,
				StartTime: start,
			}

			// Record metrics
			prometheus.RecordMetric()

			return request
		}
	}
}
