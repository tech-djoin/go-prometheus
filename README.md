# go-prometheus
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go Prometheus is a package that allows you to integrate Prometheus, a popular open-source monitoring and alerting toolkit, into your golang applications. Prometheus is widely used for monitoring various aspects of software systems, including metrics, time series data, and alerting. This package uses Memory to do the client side aggregation.


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

## License
This project is licensed under the MIT License - see the [LICENSE.md](https://github.com/MarketingPipeline/README-Quotes/blob/main/LICENSE) file for details.

## Contributors
<a href="https://github.com/tech-djoin/go-prometheus/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=tech-djoin/go-prometheus" />
</a>