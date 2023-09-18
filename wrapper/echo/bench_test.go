package echo

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func BenchmarkWithoutCollector(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.GET("/", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "Hello, world!")
	})
	e.Start(":8080")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}

func BenchmarkWithCollector(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.GET("/", func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "Hello, world!")
	})
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.Start(":8080")
	e.Use(MetricCollector())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}

type ResponseComplex struct {
	RandomInt64   int64   `json:"random_int_64"`
	RandomInt32   int32   `json:"random_int_32"`
	RandomInt     int     `json:"random_int"`
	RandomString  string  `json:"random_string"`
	RandomFloat64 float64 `json:"random_float_64"`
	RandomFloat32 float32 `json:"random_float_32"`
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func BenchmarkRandWithoutCollector(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.GET("/", func(ctx echo.Context) error {
		responseComplex := ResponseComplex{
			RandomInt64:   rand.Int63(),
			RandomInt32:   rand.Int31(),
			RandomInt:     rand.Int(),
			RandomString:  randomString(64),
			RandomFloat64: rand.Float64(),
			RandomFloat32: rand.Float32(),
		}
		return ctx.JSON(http.StatusOK, responseComplex)
	})
	e.Start(":8080")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}

func BenchmarkRandWithCollector(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.GET("/", func(ctx echo.Context) error {
		responseComplex := ResponseComplex{
			RandomInt64:   rand.Int63(),
			RandomInt32:   rand.Int31(),
			RandomInt:     rand.Int(),
			RandomString:  randomString(64),
			RandomFloat64: rand.Float64(),
			RandomFloat32: rand.Float32(),
		}
		return ctx.JSON(http.StatusOK, responseComplex)
	})
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.Start(":8080")
	e.Use(MetricCollector())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}

func externalAPICall() (string, error) {
	// Replace this URL with the actual URL of the external API
	apiURL := "https://dummyjson.com/products"

	// Make a GET request to the external API
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if the API call was successful (200 OK)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	// Read the response body and return the data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func BenchmarkGetAPIWithoutCollector(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.Start(":8080")

	e.GET("/external", func(ctx echo.Context) error {
		data, err := externalAPICall()
		if err != nil {
			return ctx.String(http.StatusInternalServerError, "Failed to fetch data from external API")
		}
		return ctx.String(http.StatusOK, data)
	})

	req := httptest.NewRequest(http.MethodGet, "/external", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}

func BenchmarkGetAPIWithCollector(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.GET("/external", func(ctx echo.Context) error {
		data, err := externalAPICall()
		if err != nil {
			return ctx.String(http.StatusInternalServerError, "Failed to fetch data from external API")
		}
		return ctx.String(http.StatusOK, data)
	})
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	req := httptest.NewRequest(http.MethodGet, "/external", nil)
	rec := httptest.NewRecorder()

	e.Start(":8080")
	e.Use(MetricCollector())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(rec, req)
	}
}
