package proxy

import (
	"time"

	"github.com/duyet/gaxy/pkg/config"
	"github.com/duyet/gaxy/pkg/errors"
	"github.com/valyala/fasthttp"
)

// Client is an enhanced HTTP client with retry logic
type Client struct {
	client     *fasthttp.Client
	retryCount int
	retryDelay time.Duration
}

// NewClient creates a new proxy client with the given configuration
func NewClient(cfg *config.Config) *Client {
	client := &fasthttp.Client{
		MaxConnsPerHost:     cfg.UpstreamMaxConns,
		MaxIdleConnDuration: 90 * time.Second,
		ReadTimeout:         cfg.UpstreamTimeout,
		WriteTimeout:        cfg.UpstreamTimeout,
		MaxConnWaitTimeout:  5 * time.Second,
	}

	return &Client{
		client:     client,
		retryCount: cfg.UpstreamRetryCount,
		retryDelay: cfg.UpstreamRetryDelay,
	}
}

// Do performs an HTTP request with retry logic
func (c *Client) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	var lastErr error

	for attempt := 0; attempt <= c.retryCount; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}

		err := c.client.Do(req, resp)
		if err == nil {
			// Success
			return nil
		}

		lastErr = err

		// Don't retry on certain errors
		if !shouldRetry(err) {
			break
		}
	}

	return errors.UpstreamError("upstream request failed after retries", lastErr)
}

// DoTimeout performs an HTTP request with a custom timeout
func (c *Client) DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	return c.client.DoTimeout(req, resp, timeout)
}

// shouldRetry determines if an error is retryable
func shouldRetry(err error) bool {
	// Retry on timeout and temporary errors
	// Don't retry on permanent errors like invalid URL, etc.
	if err == fasthttp.ErrTimeout {
		return true
	}
	if err == fasthttp.ErrConnectionClosed {
		return true
	}
	// Add more retryable errors as needed
	return false
}
