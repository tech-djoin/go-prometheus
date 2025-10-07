package goprometheus

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// RedisCollector exposes metrics stored by RedisBackend through the Prometheus registry.
type RedisCollector struct {
	backend     *RedisBackend
	requestDesc *prometheus.Desc
	latencyDesc *prometheus.Desc
}

// NewRedisCollectorFromBackend builds a collector that reads metrics from the provided Redis backend.
func NewRedisCollectorFromBackend(rb *RedisBackend) *RedisCollector {
	return &RedisCollector{
		backend: rb,
		requestDesc: prometheus.NewDesc(
			"app_http_request_totals",
			"The total number of application request http",
			[]string{"method", "path", "code"},
			nil,
		),
		latencyDesc: prometheus.NewDesc(
			"app_http_request_latency_seconds",
			"Latency of HTTP requests.",
			[]string{"method", "path"},
			nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *RedisCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.requestDesc
	ch <- c.latencyDesc
}

// Collect implements prometheus.Collector.
func (c *RedisCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := c.backend.ctx

	// Counters
	if values, err := c.backend.client.HGetAll(ctx, c.backend.key("http_requests")).Result(); err == nil {
		for field, rawValue := range values {
			labels := strings.Split(field, "|")
			if len(labels) != 3 {
				continue
			}
			count, err := strconv.ParseFloat(rawValue, 64)
			if err != nil {
				continue
			}
			method, err := decodePart(labels[0])
			if err != nil {
				continue
			}
			path, err := decodePart(labels[1])
			if err != nil {
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.requestDesc, prometheus.CounterValue, count, method, path, labels[2])
		}
	}

	// Histogram data
	totalCounts, errCount := c.backend.client.HGetAll(ctx, c.backend.key("http_latency", "count")).Result()
	sumValues, errSum := c.backend.client.HGetAll(ctx, c.backend.key("http_latency", "sum")).Result()
	bucketValues, errBuckets := c.backend.client.HGetAll(ctx, c.backend.key("http_latency", "buckets")).Result()
	if errCount != nil || errSum != nil || errBuckets != nil {
		return
	}

	bucketMap := make(map[string]map[float64]uint64)
	for field, rawValue := range bucketValues {
		parts := strings.Split(field, "|")
		if len(parts) < 3 {
			continue
		}
		bucketLabel := parts[len(parts)-1]
		base := strings.Join(parts[:len(parts)-1], "|")
		if bucketLabel == "+Inf" {
			// +Inf bucket accounted via count; skip explicit bucket entry.
			continue
		}
		upperBound, err := strconv.ParseFloat(bucketLabel, 64)
		if err != nil {
			continue
		}
		bucketCount, err := strconv.ParseUint(rawValue, 10, 64)
		if err != nil {
			continue
		}
		bucketSeries := bucketMap[base]
		if bucketSeries == nil {
			bucketSeries = make(map[float64]uint64, len(c.backend.buckets))
			bucketMap[base] = bucketSeries
		}
		bucketSeries[upperBound] = bucketCount
	}

	for base, rawCount := range totalCounts {
		count, err := strconv.ParseUint(rawCount, 10, 64)
		if err != nil {
			continue
		}
		sumRaw := sumValues[base]
		sum, err := strconv.ParseFloat(sumRaw, 64)
		if err != nil {
			continue
		}

		parts := strings.Split(base, "|")
		if len(parts) != 2 {
			continue
		}
		method, err := decodePart(parts[0])
		if err != nil {
			continue
		}
		path, err := decodePart(parts[1])
		if err != nil {
			continue
		}

		buckets := bucketMap[base]
		if buckets == nil {
			buckets = make(map[float64]uint64, len(c.backend.buckets))
		}
		// Ensure zero values for configured buckets are present.
		for _, bound := range c.backend.buckets {
			if _, ok := buckets[bound]; !ok {
				buckets[bound] = 0
			}
		}

		metric := prometheus.MustNewConstHistogram(c.latencyDesc, count, sum, buckets, method, path)
		ch <- metric
	}
}

func decodePart(value string) (string, error) {
	if value == "" {
		return value, nil
	}
	return url.PathUnescape(value)
}
