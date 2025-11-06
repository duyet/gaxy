package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/valyala/fasthttp"
)

var (
	proxyClient = &fasthttp.Client{}

	// googleDomains contains all Google Analytics and Tag Manager domains to be replaced
	googleDomains = []string{
		"ssl.google-analytics.com",
		"www.google-analytics.com",
		"google-analytics.com",
		"www.googletagmanager.com",
		"googletagmanager.com",
	}
)

func main() {
	var config = LoadConfig()
	var app = Setup(config)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", config.Port)
		if err := app.Listen(fmt.Sprintf(":%s", config.Port)); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-quit
	log.Println("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// Setup creates and configures a fiber app with all routes and middleware
func Setup(config Config) *fiber.App {
	app := fiber.New()

	// Config object
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("config", config)
		return c.Next()
	})

	// CORS
	app.Use(cors.New())

	// Logger
	app.Use(logger.New())

	// Handler
	if config.RoutePrefix != "" {
		subRoute := app.Group(config.RoutePrefix)
		subRoute.Get("/ping", pingHandler)
		subRoute.All("/*", handleRequestAndRedirect)
	}
	app.Get("/ping", pingHandler)
	app.All("/*", handleRequestAndRedirect)

	return app
}

// pingHandler handles health check requests and returns "pong"
func pingHandler(c *fiber.Ctx) error {
	return c.Send([]byte("pong"))
}

// handleRequestAndRedirect proxies incoming requests to Google Analytics/Tag Manager
// It handles URL rewriting, request preparation, and response post-processing
func handleRequestAndRedirect(c *fiber.Ctx) error {
	config := c.Locals("config").(Config)

	upstreamReq := fasthttp.AcquireRequest()
	upstreamResp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(upstreamReq)
	defer fasthttp.ReleaseResponse(upstreamResp)

	c.Request().CopyTo(upstreamReq)

	// Trim prefix
	reqURI := string(c.Request().RequestURI())
	if config.RoutePrefix != "" && strings.HasPrefix(reqURI, config.RoutePrefix+"/") {
		reqURI = strings.TrimPrefix(reqURI, config.RoutePrefix)
		upstreamReq.SetRequestURI(reqURI)
	}

	// Parse and set upstream URL
	upstreamURL, err := url.Parse(config.GoogleOrigin)
	if err != nil {
		log.Printf("Error parsing Google origin URL: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Invalid upstream URL configuration")
	}
	upstreamReq.SetHost(upstreamURL.Host)
	upstreamReq.URI().SetScheme(upstreamURL.Scheme)

	// Prepare request
	prepareRequest(upstreamReq, c)
	log.Printf("GET %s -> making request to %s", c.Params("*"), upstreamReq.URI().FullURI())

	// Start request to dest URL
	if err := proxyClient.Do(upstreamReq, upstreamResp); err != nil {
		return err
	}

	// Post process the response
	if err := postprocessResponse(upstreamResp, c); err != nil {
		return err
	}

	return nil
}

// prepareRequest prepares the upstream request by injecting headers and parameters
func prepareRequest(upstreamReq *fasthttp.Request, c *fiber.Ctx) {
	config := c.Locals("config").(Config)

	// Inject headers as query parameters
	if config.InjectParamsFromReqHeaders != "" {
		for _, name := range strings.Split(config.InjectParamsFromReqHeaders, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}

			// Handle header renaming: [HEADER_NAME]__[NEW_NAME]
			headerName := name
			paramName := name
			if strings.Contains(name, "__") {
				parts := strings.SplitN(name, "__", 2)
				headerName = parts[0]
				paramName = parts[1]
			}

			val := c.Get(headerName)
			if val != "" {
				upstreamReq.URI().QueryArgs().Add(paramName, val)
				log.Printf("Added %s=%s to query string\n", paramName, val)
			}
		}
	}

	// Skip specified parameters
	if config.SkipParamsFromReqHeaders != "" {
		for _, name := range strings.Split(config.SkipParamsFromReqHeaders, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				upstreamReq.URI().QueryArgs().Del(name)
				log.Printf("Removed %s from query string", name)
			}
		}
	}

	// Always inject IP and User-Agent
	upstreamReq.URI().QueryArgs().Add("uip", c.IP())
	upstreamReq.URI().QueryArgs().Add("ua", c.Get("User-Agent"))
}

// postprocessResponse processes the upstream response before sending it to the client
// It replaces Google domains with the current host and adds custom headers
func postprocessResponse(upstreamResp *fasthttp.Response, c *fiber.Ctx) error {
	config := c.Locals("config").(Config)

	// Add header
	upstreamResp.Header.Add("x-proxy-by", "gaxy")

	bodyString, err := GetBodyString(upstreamResp)
	if err != nil {
		return err
	}

	var contentType = string(upstreamResp.Header.ContentType())
	if strings.HasPrefix(contentType, "text/javascript") || strings.HasPrefix(contentType, "application/javascript") {
		currentHost := getGaxyHostName(c)

		for _, domain := range googleDomains {
			bodyString = strings.ReplaceAll(bodyString, domain, currentHost+config.RoutePrefix)
		}
	}

	c.Response().SetBodyString(bodyString)
	c.Response().Header.SetContentType(string(upstreamResp.Header.ContentType()))
	c.Response().SetStatusCode(upstreamResp.StatusCode())

	return nil
}

// GetBodyString extracts the body from a fasthttp.Response, handling various compression formats
// It supports gzip, brotli, and deflate compression
func GetBodyString(r *fasthttp.Response) (string, error) {
	var body []byte
	var err error

	contentEncoding := string(r.Header.Peek("Content-Encoding"))
	switch contentEncoding {
	case "gzip":
		body, err = r.BodyGunzip()
	case "br":
		body, err = r.BodyUnbrotli()
	case "deflate":
		body, err = r.BodyInflate()
	default:
		body = r.Body()
	}
	if err != nil {
		return "", err
	}

	bodyString := string(body)
	return bodyString, nil
}

// getGaxyHostName returns the host name to use for domain replacement
// It checks for X-Forwarded-Host header first (for reverse proxy setups) and falls back to the request host
func getGaxyHostName(c *fiber.Ctx) string {
	if host := c.Get("X-Forwarded-Host", ""); host != "" {
		return host
	}

	return string(c.Request().URI().Host())
}
