package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	echoprometheus "github.com/tech-djoin/go-prometheus/wrapper/echo"
)

func main() {
	e := echo.New()
	e.Use(echoprometheus.MetricCollector())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "Hello, world!")
	})
	e.Start(":8080")
}
