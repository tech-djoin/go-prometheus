package fiber

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	goprometheus "github.com/tech-djoin/go-prometheus"
)

// MetricCollector is a middleware function for collecting metrics in Fiber.
//
// It measures the time taken to process the request and records
// relevant metrics including the HTTP request and latency.
func MetricCollector() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// copy safe by this issue
		// https://github.com/prometheus/client_golang/issues/1429#issuecomment-1925737336
		method := strings.Clone(c.Method())

		// Proceed with the next handler
		err := c.Next()

		// initialize params struct
		prometheus := goprometheus.Prometheus{
			Method:    method,
			Path:      c.Route().Path,
			Code:      c.Response().StatusCode(),
			StartTime: start,
		}

		// Record metrics
		prometheus.RecordMetric()

		return err
	}
}
