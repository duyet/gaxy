package main

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/duyet/gaxy/pkg/config"
	"github.com/duyet/gaxy/pkg/handler"
	"github.com/duyet/gaxy/pkg/logger"
	"github.com/duyet/gaxy/pkg/metrics"
	"github.com/duyet/gaxy/pkg/proxy"
	"github.com/duyet/gaxy/pkg/ratelimit"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestApp() (*config.Config, *fiber.App) {
	cfg := &config.Config{
		Port:                    "3000",
		GoogleOrigin:            "https://www.google-analytics.com",
		RoutePrefix:             "",
		CacheEnabled:            true,
		CacheTTL:                5 * time.Minute,
		CacheMaxSize:            100 * 1024 * 1024,
		RateLimitEnabled:        false,
		LogLevel:                "info",
		LogFormat:               "json",
		MetricsEnabled:          true,
		MetricsPath:             "/metrics",
		EnableCORS:              true,
		EnableSecurityHeaders:   true,
		UpstreamTimeout:         10 * time.Second,
		UpstreamMaxIdleConns:    100,
		UpstreamMaxConns:        100,
		UpstreamRetryCount:      2,
		UpstreamRetryDelay:      100 * time.Millisecond,
		ReadTimeout:             30 * time.Second,
		WriteTimeout:            30 * time.Second,
		ShutdownTimeout:         10 * time.Second,
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	m := metrics.New()
	proxySvc := proxy.NewService(cfg, m, log)
	h := handler.New(cfg, proxySvc, m, log)
	var limiter *ratelimit.Limiter

	app := Setup(cfg, h, m, limiter, log)
	return cfg, app
}

func TestServer(t *testing.T) {
	_, app := setupTestApp()

	req := httptest.NewRequest("GET", "/ping", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err, "request should not fail")
	assert.Equal(t, 200, resp.StatusCode, "status code should be 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.Equal(t, "pong", string(body), "body should be 'pong'")
}

func TestGAJS(t *testing.T) {
	_, app := setupTestApp()

	req := httptest.NewRequest("GET", "/ga.js", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err, "request should not fail")
	assert.Equal(t, 200, resp.StatusCode, "status code should be 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.NotEmpty(t, string(body), "body should not be empty")
	assert.Contains(t, string(body), "google", "body should contain 'google' keyword")
}

func TestHealthEndpoint(t *testing.T) {
	_, app := setupTestApp()

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err, "request should not fail")
	assert.Equal(t, 200, resp.StatusCode, "status code should be 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.Contains(t, string(body), "healthy", "body should contain healthy status")
}

func TestMetricsEndpoint(t *testing.T) {
	_, app := setupTestApp()

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err, "request should not fail")
	assert.Equal(t, 200, resp.StatusCode, "status code should be 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "reading body should not fail")
	assert.Contains(t, string(body), "gaxy_", "body should contain gaxy metrics")
}
