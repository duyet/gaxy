package middleware

import (
	"time"

	"github.com/duyet/gaxy/pkg/config"
	"github.com/duyet/gaxy/pkg/logger"
	"github.com/duyet/gaxy/pkg/metrics"
	"github.com/duyet/gaxy/pkg/ratelimit"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !cfg.EnableSecurityHeaders {
			return c.Next()
		}

		// Prevent clickjacking
		c.Set("X-Frame-Options", "SAMEORIGIN")

		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Set("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Don't expose server version
		c.Set("X-Powered-By", "gaxy")

		return c.Next()
	}
}

// RateLimit implements per-IP rate limiting
func RateLimit(cfg *config.Config, limiter *ratelimit.Limiter, m *metrics.Metrics, log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !cfg.RateLimitEnabled {
			return c.Next()
		}

		ip := c.IP()
		if !limiter.Allow(ip) {
			m.RecordRateLimitDrop()
			log.WithField("ip", ip).Warn("Rate limit exceeded")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		}

		return c.Next()
	}
}

// Metrics tracks request metrics
func Metrics(m *metrics.Metrics) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		m.IncRequestsInFlight()
		defer m.DecRequestsInFlight()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()
		m.RecordRequest(statusCode, duration)

		return err
	}
}

// RequestID adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID already exists (from upstream proxy)
		requestID := c.Get("X-Request-ID", "")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)

		return c.Next()
	}
}

// Logger logs requests with structured logging
func Logger(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Get request ID
		requestID := c.Locals("request_id")
		if requestID == nil {
			requestID = "unknown"
		}

		// Create logger with request context
		reqLogger := log.WithFields(map[string]interface{}{
			"request_id": requestID,
			"method":     c.Method(),
			"path":       c.Path(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
		})

		// Store logger in context
		c.Locals("logger", reqLogger)

		// Log request
		reqLogger.Info("Request started")

		// Process request
		err := c.Next()

		// Log response
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()

		respLogger := reqLogger.WithFields(map[string]interface{}{
			"status":        statusCode,
			"duration_ms":   duration.Milliseconds(),
			"response_size": len(c.Response().Body()),
		})

		if err != nil {
			respLogger.WithField("error", err.Error()).Error("Request failed")
		} else if statusCode >= 500 {
			respLogger.Error("Request completed with error")
		} else if statusCode >= 400 {
			respLogger.Warn("Request completed with client error")
		} else {
			respLogger.Info("Request completed")
		}

		return err
	}
}

// Recovery recovers from panics and returns a 500 error
func Recovery(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.WithField("panic", r).Error("Panic recovered")

				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}()

		return c.Next()
	}
}

// CORS handles CORS with configurable origins
func CORS(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !cfg.EnableCORS {
			return c.Next()
		}

		origin := cfg.CORSAllowOrigins
		if origin == "" {
			origin = "*"
		}

		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}
