package handler

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/duyet/gaxy/pkg/config"
	"github.com/duyet/gaxy/pkg/logger"
	"github.com/duyet/gaxy/pkg/metrics"
	"github.com/duyet/gaxy/pkg/proxy"
	"github.com/gofiber/fiber/v2"
)

// Handler contains all HTTP handlers
type Handler struct {
	config       *config.Config
	proxyService *proxy.Service
	metrics      *metrics.Metrics
	logger       *logger.Logger
	startTime    time.Time
}

// New creates a new handler
func New(cfg *config.Config, proxySvc *proxy.Service, m *metrics.Metrics, log *logger.Logger) *Handler {
	return &Handler{
		config:       cfg,
		proxyService: proxySvc,
		metrics:      m,
		logger:       log,
		startTime:    time.Now(),
	}
}

// Ping handles health check requests
func (h *Handler) Ping(c *fiber.Ctx) error {
	return c.SendString("pong")
}

// Health provides detailed health information
func (h *Handler) Health(c *fiber.Ctx) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	uptime := time.Since(h.startTime)

	health := fiber.Map{
		"status":  "healthy",
		"version": "1.0.0",
		"uptime":  uptime.String(),
		"system": fiber.Map{
			"goroutines":    runtime.NumGoroutine(),
			"memory_alloc":  fmt.Sprintf("%d MB", memStats.Alloc/1024/1024),
			"memory_total":  fmt.Sprintf("%d MB", memStats.TotalAlloc/1024/1024),
			"memory_sys":    fmt.Sprintf("%d MB", memStats.Sys/1024/1024),
			"gc_runs":       memStats.NumGC,
		},
	}

	return c.JSON(health)
}

// Metrics exports Prometheus metrics
func (h *Handler) Metrics(c *fiber.Ctx) error {
	output := h.metrics.Export()
	c.Set("Content-Type", "text/plain; version=0.0.4")
	return c.SendString(output)
}

// Proxy handles all proxy requests
func (h *Handler) Proxy(c *fiber.Ctx) error {
	// Get logger from context if available
	log := h.logger
	if ctxLogger := c.Locals("logger"); ctxLogger != nil {
		if l, ok := ctxLogger.(*logger.Logger); ok {
			log = l
		}
	}

	// Get request URI
	reqURI := string(c.Request().RequestURI())

	// Trim route prefix if configured
	if h.config.RoutePrefix != "" && strings.HasPrefix(reqURI, h.config.RoutePrefix+"/") {
		reqURI = strings.TrimPrefix(reqURI, h.config.RoutePrefix)
	}

	// Collect headers
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	// Add IP and User-Agent to headers for injection
	headers["uip"] = c.IP()
	if ua := c.Get("User-Agent"); ua != "" {
		headers["ua"] = ua
	}

	// Get host for domain replacement
	host := h.getHostName(c)

	log.WithFields(map[string]interface{}{
		"uri":  reqURI,
		"host": host,
	}).Debug("Processing proxy request")

	// Proxy the request
	resp, err := h.proxyService.ProxyRequest(reqURI, headers, host)
	if err != nil {
		log.WithField("error", err.Error()).Error("Proxy request failed")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to proxy request",
		})
	}

	// Add custom headers
	c.Set("X-Proxy-By", "gaxy")

	// Set response
	c.Set("Content-Type", resp.ContentType)
	c.Status(resp.StatusCode)
	return c.Send(resp.Body)
}

// getHostName returns the host name for domain replacement
func (h *Handler) getHostName(c *fiber.Ctx) string {
	// Check X-Forwarded-Host for reverse proxy setups
	if host := c.Get("X-Forwarded-Host"); host != "" {
		return host
	}

	// Fallback to request host
	return string(c.Request().URI().Host())
}
