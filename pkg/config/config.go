package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/duyet/gaxy/pkg/errors"
	"github.com/kelseyhightower/envconfig"
)

// Config contains application configuration loaded from environment variables
type Config struct {
	// Server configuration
	Port            string        `env:"PORT" default:"3000"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" default:"10s"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" default:"30s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" default:"30s"`

	// Routing configuration
	RoutePrefix string `env:"ROUTE_PREFIX" default:""`

	// Upstream configuration
	GoogleOrigin           string        `env:"GOOGLE_ORIGIN" default:"https://www.google-analytics.com"`
	UpstreamTimeout        time.Duration `env:"UPSTREAM_TIMEOUT" default:"10s"`
	UpstreamMaxIdleConns   int           `env:"UPSTREAM_MAX_IDLE_CONNS" default:"100"`
	UpstreamMaxConns       int           `env:"UPSTREAM_MAX_CONNS" default:"100"`
	UpstreamRetryCount     int           `env:"UPSTREAM_RETRY_COUNT" default:"2"`
	UpstreamRetryDelay     time.Duration `env:"UPSTREAM_RETRY_DELAY" default:"100ms"`

	// Header injection configuration
	InjectParamsFromReqHeaders string `env:"INJECT_PARAMS_FROM_REQ_HEADERS" default:""`
	SkipParamsFromReqHeaders   string `env:"SKIP_PARAMS_FROM_REQ_HEADERS" default:""`

	// Cache configuration
	CacheEnabled    bool          `env:"CACHE_ENABLED" default:"true"`
	CacheTTL        time.Duration `env:"CACHE_TTL" default:"5m"`
	CacheMaxSize    int64         `env:"CACHE_MAX_SIZE" default:"104857600"` // 100MB
	CacheKeyPattern string        `env:"CACHE_KEY_PATTERN" default:"*.js"`

	// Rate limiting configuration
	RateLimitEnabled bool `env:"RATE_LIMIT_ENABLED" default:"true"`
	RateLimitRPS     int  `env:"RATE_LIMIT_RPS" default:"100"`     // requests per second per IP
	RateLimitBurst   int  `env:"RATE_LIMIT_BURST" default:"200"`   // burst allowance

	// Logging configuration
	LogLevel  string `env:"LOG_LEVEL" default:"info"`
	LogFormat string `env:"LOG_FORMAT" default:"json"` // json or text

	// Metrics configuration
	MetricsEnabled bool   `env:"METRICS_ENABLED" default:"true"`
	MetricsPath    string `env:"METRICS_PATH" default:"/metrics"`

	// Security configuration
	EnableCORS           bool   `env:"ENABLE_CORS" default:"true"`
	CORSAllowOrigins     string `env:"CORS_ALLOW_ORIGINS" default:"*"`
	EnableSecurityHeaders bool   `env:"ENABLE_SECURITY_HEADERS" default:"true"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{}
	if err := envconfig.Process("", config); err != nil {
		return nil, errors.ConfigError("failed to process environment variables", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate port
	if c.Port == "" {
		return errors.ValidationError("PORT cannot be empty")
	}

	// Validate Google origin URL
	if c.GoogleOrigin == "" {
		return errors.ValidationError("GOOGLE_ORIGIN cannot be empty")
	}

	parsedURL, err := url.Parse(c.GoogleOrigin)
	if err != nil {
		return errors.Wrap(errors.ErrorTypeValidation, "invalid GOOGLE_ORIGIN URL", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.ValidationError(fmt.Sprintf("GOOGLE_ORIGIN must use http or https scheme, got: %s", parsedURL.Scheme))
	}

	// Validate route prefix format
	if c.RoutePrefix != "" {
		if !strings.HasPrefix(c.RoutePrefix, "/") {
			return errors.ValidationError("ROUTE_PREFIX must start with /")
		}
		if strings.HasSuffix(c.RoutePrefix, "/") {
			return errors.ValidationError("ROUTE_PREFIX must not end with /")
		}
	}

	// Validate timeouts
	if c.UpstreamTimeout <= 0 {
		return errors.ValidationError("UPSTREAM_TIMEOUT must be positive")
	}
	if c.ReadTimeout <= 0 {
		return errors.ValidationError("READ_TIMEOUT must be positive")
	}
	if c.WriteTimeout <= 0 {
		return errors.ValidationError("WRITE_TIMEOUT must be positive")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.ValidationError("SHUTDOWN_TIMEOUT must be positive")
	}

	// Validate connection pool sizes
	if c.UpstreamMaxIdleConns <= 0 {
		return errors.ValidationError("UPSTREAM_MAX_IDLE_CONNS must be positive")
	}
	if c.UpstreamMaxConns <= 0 {
		return errors.ValidationError("UPSTREAM_MAX_CONNS must be positive")
	}
	if c.UpstreamMaxConns < c.UpstreamMaxIdleConns {
		return errors.ValidationError("UPSTREAM_MAX_CONNS must be >= UPSTREAM_MAX_IDLE_CONNS")
	}

	// Validate retry configuration
	if c.UpstreamRetryCount < 0 {
		return errors.ValidationError("UPSTREAM_RETRY_COUNT cannot be negative")
	}
	if c.UpstreamRetryDelay < 0 {
		return errors.ValidationError("UPSTREAM_RETRY_DELAY cannot be negative")
	}

	// Validate cache configuration
	if c.CacheEnabled {
		if c.CacheTTL <= 0 {
			return errors.ValidationError("CACHE_TTL must be positive when cache is enabled")
		}
		if c.CacheMaxSize <= 0 {
			return errors.ValidationError("CACHE_MAX_SIZE must be positive when cache is enabled")
		}
	}

	// Validate rate limit configuration
	if c.RateLimitEnabled {
		if c.RateLimitRPS <= 0 {
			return errors.ValidationError("RATE_LIMIT_RPS must be positive when rate limiting is enabled")
		}
		if c.RateLimitBurst <= 0 {
			return errors.ValidationError("RATE_LIMIT_BURST must be positive when rate limiting is enabled")
		}
	}

	// Validate log level
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		return errors.ValidationError(fmt.Sprintf("invalid LOG_LEVEL: %s (must be debug, info, warn, or error)", c.LogLevel))
	}

	// Validate log format
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return errors.ValidationError(fmt.Sprintf("invalid LOG_FORMAT: %s (must be json or text)", c.LogFormat))
	}

	return nil
}

// GetParsedGoogleOrigin returns the parsed Google origin URL
func (c *Config) GetParsedGoogleOrigin() (*url.URL, error) {
	return url.Parse(c.GoogleOrigin)
}

// GetInjectHeaders returns the list of headers to inject as query parameters
func (c *Config) GetInjectHeaders() []HeaderMapping {
	if c.InjectParamsFromReqHeaders == "" {
		return nil
	}

	var mappings []HeaderMapping
	for _, mapping := range strings.Split(c.InjectParamsFromReqHeaders, ",") {
		mapping = strings.TrimSpace(mapping)
		if mapping == "" {
			continue
		}

		var headerName, paramName string
		if strings.Contains(mapping, "__") {
			parts := strings.SplitN(mapping, "__", 2)
			headerName = parts[0]
			paramName = parts[1]
		} else {
			headerName = mapping
			paramName = mapping
		}

		mappings = append(mappings, HeaderMapping{
			HeaderName: headerName,
			ParamName:  paramName,
		})
	}

	return mappings
}

// GetSkipParams returns the list of parameters to skip
func (c *Config) GetSkipParams() []string {
	if c.SkipParamsFromReqHeaders == "" {
		return nil
	}

	var params []string
	for _, param := range strings.Split(c.SkipParamsFromReqHeaders, ",") {
		param = strings.TrimSpace(param)
		if param != "" {
			params = append(params, param)
		}
	}

	return params
}

// HeaderMapping represents a header to query parameter mapping
type HeaderMapping struct {
	HeaderName string
	ParamName  string
}
