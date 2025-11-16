package proxy

import (
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/duyet/gaxy/pkg/cache"
	"github.com/duyet/gaxy/pkg/config"
	"github.com/duyet/gaxy/pkg/errors"
	"github.com/duyet/gaxy/pkg/logger"
	"github.com/duyet/gaxy/pkg/metrics"
	"github.com/valyala/fasthttp"
)

var (
	// googleDomains contains all Google Analytics and Tag Manager domains to be replaced
	googleDomains = []string{
		"ssl.google-analytics.com",
		"www.google-analytics.com",
		"google-analytics.com",
		"www.googletagmanager.com",
		"googletagmanager.com",
	}
)

// Service handles proxy operations
type Service struct {
	config  *config.Config
	client  *Client
	cache   *cache.Cache
	metrics *metrics.Metrics
	logger  *logger.Logger
}

// NewService creates a new proxy service
func NewService(cfg *config.Config, m *metrics.Metrics, log *logger.Logger) *Service {
	var c *cache.Cache
	if cfg.CacheEnabled {
		c = cache.New(cfg.CacheTTL, cfg.CacheMaxSize)
	}

	return &Service{
		config:  cfg,
		client:  NewClient(cfg),
		cache:   c,
		metrics: m,
		logger:  log,
	}
}

// ProxyRequest proxies a request to Google Analytics/Tag Manager
func (s *Service) ProxyRequest(reqURI string, headers map[string]string, host string) (*Response, error) {
	start := time.Now()

	// Check cache first
	if s.cache != nil && s.isCacheable(reqURI) {
		cacheKey := s.getCacheKey(reqURI)
		if entry, found := s.cache.Get(cacheKey); found {
			s.logger.WithField("uri", reqURI).Debug("Cache hit")

			// Update metrics
			stats := s.cache.GetStats()
			s.metrics.UpdateCacheStats(stats.Hits, stats.Misses, stats.Evictions, stats.Size, int64(stats.EntryCount))

			return &Response{
				StatusCode:  entry.StatusCode,
				Body:        entry.Data,
				ContentType: entry.ContentType,
			}, nil
		}
		s.logger.WithField("uri", reqURI).Debug("Cache miss")
	}

	// Prepare upstream request
	upstreamReq := fasthttp.AcquireRequest()
	upstreamResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(upstreamReq)
	defer fasthttp.ReleaseResponse(upstreamResp)

	// Parse upstream URL
	upstreamURL, err := s.config.GetParsedGoogleOrigin()
	if err != nil {
		return nil, errors.ConfigError("invalid upstream URL", err)
	}

	// Build request
	upstreamReq.SetRequestURI(reqURI)
	upstreamReq.SetHost(upstreamURL.Host)
	upstreamReq.URI().SetScheme(upstreamURL.Scheme)

	// Copy headers
	for key, value := range headers {
		upstreamReq.Header.Set(key, value)
	}

	// Inject configured headers as query parameters
	for _, mapping := range s.config.GetInjectHeaders() {
		if val, ok := headers[mapping.HeaderName]; ok && val != "" {
			upstreamReq.URI().QueryArgs().Add(mapping.ParamName, val)
			s.logger.WithFields(map[string]interface{}{
				"header": mapping.HeaderName,
				"param":  mapping.ParamName,
				"value":  val,
			}).Debug("Injected header as query param")
		}
	}

	// Skip configured parameters
	for _, param := range s.config.GetSkipParams() {
		upstreamReq.URI().QueryArgs().Del(param)
		s.logger.WithField("param", param).Debug("Removed query param")
	}

	s.logger.WithField("upstream_uri", upstreamReq.URI().String()).Debug("Proxying request")

	// Make request
	err = s.client.Do(upstreamReq, upstreamResp)

	// Record upstream metrics
	duration := time.Since(start)
	statusCode := upstreamResp.StatusCode()
	s.metrics.RecordUpstreamRequest(statusCode, duration, err != nil)

	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Upstream request failed")
		return nil, errors.UpstreamError("failed to proxy request", err)
	}

	// Extract response body
	bodyString, err := s.getBodyString(upstreamResp)
	if err != nil {
		return nil, errors.ProxyError("failed to read response body", err)
	}

	contentType := string(upstreamResp.Header.ContentType())

	// Post-process response (replace Google domains)
	if s.isJavaScript(contentType) {
		currentHost := host
		for _, domain := range googleDomains {
			bodyString = strings.ReplaceAll(bodyString, domain, currentHost+s.config.RoutePrefix)
		}
	}

	bodyBytes := []byte(bodyString)

	// Cache if applicable
	if s.cache != nil && s.isCacheable(reqURI) && statusCode == 200 {
		cacheKey := s.getCacheKey(reqURI)
		s.cache.Set(cacheKey, bodyBytes, contentType, statusCode)
		s.logger.WithField("cache_key", cacheKey).Debug("Cached response")

		// Update cache metrics
		stats := s.cache.GetStats()
		s.metrics.UpdateCacheStats(stats.Hits, stats.Misses, stats.Evictions, stats.Size, int64(stats.EntryCount))
	}

	return &Response{
		StatusCode:  statusCode,
		Body:        bodyBytes,
		ContentType: contentType,
	}, nil
}

// isCacheable determines if a URI should be cached
func (s *Service) isCacheable(uri string) bool {
	if !s.config.CacheEnabled {
		return false
	}

	// Extract path from URI
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return false
	}

	path := parsedURI.Path

	// Check if path matches cache pattern
	matched, _ := filepath.Match(s.config.CacheKeyPattern, filepath.Base(path))
	return matched
}

// getCacheKey generates a cache key from a URI
func (s *Service) getCacheKey(uri string) string {
	return uri
}

// isJavaScript checks if content type is JavaScript
func (s *Service) isJavaScript(contentType string) bool {
	return strings.HasPrefix(contentType, "text/javascript") ||
		strings.HasPrefix(contentType, "application/javascript") ||
		strings.HasPrefix(contentType, "application/x-javascript")
}

// getBodyString extracts the body from a fasthttp.Response, handling compression
func (s *Service) getBodyString(r *fasthttp.Response) (string, error) {
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

	return string(body), nil
}

// Response represents a proxy response
type Response struct {
	StatusCode  int
	Body        []byte
	ContentType string
}
