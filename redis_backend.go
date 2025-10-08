package goprometheus

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

// RedisBackendOptions configures a Redis-backed metrics store.
type RedisBackendOptions struct {
	Client    redis.Cmdable
	Context   context.Context
	KeyPrefix string
	Buckets   []float64
	OnError   func(error)
}

// RedisBackend persists metrics in Redis hashes.
type RedisBackend struct {
	client    redis.Cmdable
	ctx       context.Context
	keyPrefix string
	buckets   []float64
	onError   func(error)
}

// NewRedisBackend creates a new Redis metrics backend.
func NewRedisBackend(opts RedisBackendOptions) (*RedisBackend, error) {
	client := opts.Client
	if client == nil {
		var err error
		client, err = redisClientFromEnv()
		if err != nil {
			return nil, err
		}
		if client == nil {
			return nil, fmt.Errorf("goprometheus: redis client cannot be nil")
		}
	}

	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	buckets := opts.Buckets
	if len(buckets) == 0 {
		buckets = append([]float64(nil), DefaultLatencyBuckets...)
	} else {
		buckets = append([]float64(nil), buckets...)
		sort.Float64s(buckets)
	}

	return &RedisBackend{
		client:    client,
		ctx:       ctx,
		keyPrefix: opts.KeyPrefix,
		buckets:   buckets,
		onError:   opts.OnError,
	}, nil
}

// UseRedisBackend builds and installs a Redis-backed metrics backend using the
// provided options.
func UseRedisBackend(opts RedisBackendOptions) (*RedisBackend, error) {
	backend, err := NewRedisBackend(opts)
	if err != nil {
		return nil, err
	}
	UseBackend(backend)
	return backend, nil
}

func (r *RedisBackend) RecordHttpRequest(method string, path string, code int) {
	field := strings.Join([]string{encodePart(method), encodePart(path), strconv.Itoa(code)}, "|")
	r.handleError(r.client.HIncrBy(r.ctx, r.key("http_requests"), field, 1).Err())
}

func (r *RedisBackend) RecordLatency(method string, path string, elapsedSeconds float64) {
	baseField := strings.Join([]string{encodePart(method), encodePart(path)}, "|")
	r.handleError(r.client.HIncrBy(r.ctx, r.key("http_latency", "count"), baseField, 1).Err())
	r.handleError(r.client.HIncrByFloat(r.ctx, r.key("http_latency", "sum"), baseField, elapsedSeconds).Err())

	for _, bucket := range r.buckets {
		if elapsedSeconds <= bucket {
			field := strings.Join([]string{baseField, strconv.FormatFloat(bucket, 'f', -1, 64)}, "|")
			r.handleError(r.client.HIncrBy(r.ctx, r.key("http_latency", "buckets"), field, 1).Err())
		}
	}

	r.handleError(r.client.HIncrBy(r.ctx, r.key("http_latency", "buckets"), baseField+"|+Inf", 1).Err())
}

func (r *RedisBackend) key(parts ...string) string {
	if len(parts) == 0 {
		return r.keyPrefix
	}
	if r.keyPrefix == "" {
		return strings.Join(parts, ":")
	}
	return r.keyPrefix + ":" + strings.Join(parts, ":")
}

func (r *RedisBackend) handleError(err error) {
	if err != nil && r.onError != nil {
		r.onError(err)
	}
}

func encodePart(value string) string {
	if value == "" {
		return value
	}
	return url.PathEscape(value)
}

const (
	envRedisAddr     = "GOPROM_REDIS_ADDR"
	envRedisHost     = "GOPROM_REDIS_HOST"
	envRedisPort     = "GOPROM_REDIS_PORT"
	envRedisUsername = "GOPROM_REDIS_USERNAME"
	envRedisPassword = "GOPROM_REDIS_PASSWORD"
)

func redisClientFromEnv() (redis.Cmdable, error) {
	addr, err := redisAddrFromEnv()
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(os.Getenv(envRedisUsername))
	password := os.Getenv(envRedisPassword)

	opts := &redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
	}

	return redis.NewClient(opts), nil
}

func redisAddrFromEnv() (string, error) {
	if addr := strings.TrimSpace(os.Getenv(envRedisAddr)); addr != "" {
		return addr, nil
	}
	host := strings.TrimSpace(os.Getenv(envRedisHost))
	if host == "" {
		host = "127.0.0.1"
	}
	port := strings.TrimSpace(os.Getenv(envRedisPort))
	if port == "" {
		port = "6379"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return "", fmt.Errorf("goprometheus: invalid %s value %q: %w", envRedisPort, port, err)
	}
	return net.JoinHostPort(host, port), nil
}
