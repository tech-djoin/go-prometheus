package gin

import (
	"time"

	"github.com/gin-gonic/gin"
	goprometheus "github.com/tech-djoin/go-prometheus"
)

// MetricCollector is a middleware function for collecting metrics in an Gin framework application.
//
// This function returns a middleware function that wraps the next handler function in the Gin
// framework's middleware chain. It measures the time taken to process the request and records
// relevant metrics including the HTTP request and latency.
func MetricCollector() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// initialize params struct
		prometheus := goprometheus.Prometheus{
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Code:      c.Writer.Status(),
			StartTime: start,
		}

		// Record metrics
		prometheus.RecordMetric()
	}
}
