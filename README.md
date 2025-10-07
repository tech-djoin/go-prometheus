# go-prometheus
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go Prometheus is a package that allows you to integrate Prometheus, a popular open-source monitoring and alerting toolkit, into your golang applications. Prometheus is widely used for monitoring various aspects of software systems, including metrics, time series data, and alerting. By default this package aggregates metrics in memory, and it can also persist metrics to Redis when you need to share state across multiple processes.


## Installation
Get Go Prometheus package on your project:

```bash
go get github.com/tech-djoin/go-prometheus
```

## Usage
This packages provides a middleware using Echo which can be added as a global middleware or as a single route.

```go
// in server file or anywhere middleware should be registered
e.Use(goprometheus.MetricCollector())
```

```go
// in route file or anywhere route should be registered
router.Echo.GET("api/v1/posts", handler, goprometheus.MetricCollector())
```

## Exporting Metrics
To exposes all metrics gathered by collectors, you need a route to access all metrics.

```go
router.Echo.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
```

## Using Redis for Metric Storage
If you want to aggregate metrics across several application instances you can switch to the Redis-backed storage. Provide a Redis client (from `github.com/redis/go-redis/v9`) and install the backend before you start emitting metrics:

```go
import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	goprometheus "github.com/tech-djoin/go-prometheus"
)

func init() {
	backend, err := goprometheus.NewRedisBackend(goprometheus.RedisBackendOptions{
		KeyPrefix: "app_metrics", // optional; defaults to no prefix
	})
	if err != nil {
		log.Fatalf("new redis backend: %v", err)
	}

	// Use Redis for recording metrics and register a collector to expose them.
	goprometheus.UseBackend(backend)
	prometheus.MustRegister(goprometheus.NewRedisCollectorFromBackend(backend))
}
```

When `RedisBackendOptions.Client` is omitted, the package reads its connection details from the following environment variables:

- `GOPROM_REDIS_ADDR` (optional) – full `host:port` value when you want to specify the address directly.
- `GOPROM_REDIS_HOST` and `GOPROM_REDIS_PORT` – used when `GOPROM_REDIS_ADDR` is not set (default `127.0.0.1:6379`).
- `GOPROM_REDIS_USERNAME` and `GOPROM_REDIS_PASSWORD` – used when your Redis deployment requires credentials.

Metrics are stored in a set of Redis hashes keyed by the prefix you provide. The default latency histogram buckets (`goprometheus.DefaultLatencyBuckets`) are reused unless you override them in `RedisBackendOptions`.

## License
This project is licensed under the MIT License - see the [LICENSE.md](https://github.com/MarketingPipeline/README-Quotes/blob/main/LICENSE) file for details.

## Contributors
<a href="https://github.com/tech-djoin/go-prometheus/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=tech-djoin/go-prometheus" />
</a>
