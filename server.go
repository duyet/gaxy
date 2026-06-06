package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/duyet/gaxy/pkg/config"
	"github.com/duyet/gaxy/pkg/handler"
	"github.com/duyet/gaxy/pkg/logger"
	"github.com/duyet/gaxy/pkg/metrics"
	"github.com/duyet/gaxy/pkg/middleware"
	"github.com/duyet/gaxy/pkg/proxy"
	"github.com/duyet/gaxy/pkg/ratelimit"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Check for healthcheck flag
	for _, arg := range os.Args[1:] {
		if arg == "--healthcheck" || arg == "--health-check" {
			if err := runHealthCheck(cfg.Port); err != nil {
				fmt.Printf("Health check failed: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	log.Info("Starting gaxy...")

	// Initialize metrics
	m := metrics.New()

	// Initialize rate limiter
	var limiter *ratelimit.Limiter
	if cfg.RateLimitEnabled {
		limiter = ratelimit.New(cfg.RateLimitRPS, cfg.RateLimitBurst)
		log.WithFields(map[string]interface{}{
			"rps":   cfg.RateLimitRPS,
			"burst": cfg.RateLimitBurst,
		}).Info("Rate limiting enabled")
	}

	// Initialize proxy service
	proxySvc := proxy.NewService(cfg, m, log)

	// Initialize handlers
	h := handler.New(cfg, proxySvc, m, log)

	// Setup app
	app := Setup(cfg, h, m, limiter, log)

	// Start server in a goroutine
	go func() {
		log.WithField("port", cfg.Port).Info("Server starting")
		if err := app.Listen(fmt.Sprintf(":%s", cfg.Port)); err != nil {
			log.WithField("error", err.Error()).Error("Server error")
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-quit
	log.Info("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.WithField("error", err.Error()).Error("Server forced to shutdown")
		os.Exit(1)
	}

	log.Info("Server exited gracefully")
}

// Setup creates and configures a fiber app with all routes and middleware
func Setup(cfg *config.Config, h *handler.Handler, m *metrics.Metrics, limiter *ratelimit.Limiter, log *logger.Logger) *fiber.App {
	// Create app with custom config
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Global middleware
	app.Use(middleware.Recovery(log))
	app.Use(middleware.RequestID())
	app.Use(middleware.CORS(cfg))
	app.Use(middleware.SecurityHeaders(cfg))
	app.Use(middleware.Logger(log))
	app.Use(middleware.Metrics(m))

	if limiter != nil {
		app.Use(middleware.RateLimit(cfg, limiter, m, log))
	}

	// Health and metrics endpoints
	app.Get("/ping", h.Ping)
	app.Get("/health", h.Health)

	if cfg.MetricsEnabled {
		app.Get(cfg.MetricsPath, h.Metrics)
	}

	// Proxy routes
	if cfg.RoutePrefix != "" {
		subRoute := app.Group(cfg.RoutePrefix)
		subRoute.All("/*", h.Proxy)
	}
	app.All("/*", h.Proxy)

	return app
}

func runHealthCheck(port string) error {
	url := fmt.Sprintf("http://127.0.0.1:%s/ping", port)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
