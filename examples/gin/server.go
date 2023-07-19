package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	ginpromtheus "github.com/tech-djoin/go-prometheus/wrapper/gin"
)

func main() {
	r := gin.Default()
	r.Use(ginpromtheus.MetricCollector())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, world!")
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Run()
}
