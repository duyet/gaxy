# CLAUDE.md - Gaxy Development Guide

This document serves as the guiding philosophy and technical reference for AI assistants working on the Gaxy project.

## Project Philosophy

### Think Different
Gaxy is not just a proxy—it's a production-grade analytics infrastructure. Every decision should prioritize:
1. **Reliability** - Code that works under pressure
2. **Observability** - Know what's happening, always
3. **Security** - Protect the infrastructure and users
4. **Performance** - Fast is a feature
5. **Maintainability** - Code that welcomes change

### Obsess Over Details
- Every function name should be self-documenting
- Every error should provide context for debugging
- Every configuration option should have validation
- Every package should have a single, clear responsibility
- Every line of code should justify its existence

### Simplicity Through Sophistication
- Use advanced techniques to achieve simple interfaces
- Hide complexity in well-abstracted packages
- Prefer explicit over implicit
- Fail fast with clear error messages

## Architecture Principles

### Clean Architecture
```
Presentation Layer (handlers)
    ↓
Business Logic (services)
    ↓
Infrastructure (proxy, cache, logger)
```

**Key Rules:**
- Handlers don't contain business logic
- Services don't know about HTTP
- Infrastructure packages are self-contained
- Dependencies point inward (handlers → services → infrastructure)

### Package Organization

```
pkg/
├── cache/       - Intelligent caching with LRU + TTL
├── config/      - Configuration with validation
├── errors/      - Typed errors with context
├── handler/     - HTTP endpoint handlers
├── logger/      - Structured logging
├── metrics/     - Prometheus metrics collection
├── middleware/  - HTTP middleware stack
├── proxy/       - Upstream communication
└── ratelimit/   - Per-IP rate limiting
```

### Error Handling Philosophy

**Always:**
- Use typed errors from `pkg/errors`
- Include context with `.WithContext()`
- Log errors at the source
- Propagate errors up the stack
- Return user-friendly messages to clients

**Never:**
- Swallow errors silently
- Use generic error messages
- Leak internal details to clients
- Panic in production code (except startup validation)

Example:
```go
// Good
if err := validateURL(cfg.GoogleOrigin); err != nil {
    return errors.Wrap(errors.ErrorTypeConfig, "invalid Google origin", err)
}

// Bad
if err != nil {
    log.Println("error")
    return err
}
```

### Logging Standards

**Use structured logging:**
```go
log.WithFields(map[string]interface{}{
    "request_id": requestID,
    "upstream":   upstreamURL,
    "duration":   duration,
}).Info("Request proxied successfully")
```

**Log Levels:**
- `DEBUG` - Detailed diagnostic information (disabled in production)
- `INFO` - General operational messages
- `WARN` - Unexpected but recoverable situations
- `ERROR` - Errors that need attention

**Always include:**
- Request ID for correlation
- Relevant context (user, resource, action)
- Duration for performance tracking

### Metrics Philosophy

**Track what matters:**
- Request rates and latencies (percentiles: p50, p95, p99)
- Cache hit/miss rates
- Upstream health and response times
- Error rates by type
- Resource usage (connections, memory)

**Naming convention:**
```
gaxy_{component}_{metric}_{unit}
Examples:
- gaxy_requests_total
- gaxy_request_duration_seconds
- gaxy_cache_hits_total
- gaxy_upstream_errors_total
```

### Configuration Management

**Validation Rules:**
1. Validate ALL configuration at startup
2. Use typed config structs
3. Provide sensible defaults
4. Document every environment variable
5. Fail fast on invalid config (don't start the server)

**Example:**
```go
// All configs must implement Validate()
func (c *Config) Validate() error {
    if c.Port == "" {
        return errors.ValidationError("PORT cannot be empty")
    }
    // ... more validation
    return nil
}
```

### Performance Guidelines

**Caching:**
- Cache immutable assets aggressively (analytics.js, gtag.js)
- Use TTL for cache invalidation
- Track cache statistics for tuning
- Set reasonable size limits

**Connection Management:**
- Reuse connections (connection pooling)
- Set appropriate timeouts
- Limit concurrent connections
- Handle connection failures gracefully

**Memory Management:**
- Avoid unbounded growth (use size limits)
- Clean up stale data periodically
- Use sync.Pool for frequently allocated objects
- Profile in production

### Security Best Practices

**Input Validation:**
- Validate all external input
- Sanitize URLs and headers
- Enforce size limits
- Use allowlists over blocklists

**Rate Limiting:**
- Enable by default
- Per-IP limits to prevent abuse
- Configurable thresholds
- Log rate limit violations

**Headers:**
- Always set security headers
- CORS with explicit origins in production
- No sensitive data in logs or error messages
- Use HTTPS in production

## Code Style Guide

### Naming Conventions
- **Packages:** lowercase, single word (e.g., `cache`, `proxy`)
- **Types:** PascalCase (e.g., `ProxyService`, `CacheEntry`)
- **Functions:** PascalCase for exported, camelCase for private
- **Constants:** SCREAMING_SNAKE_CASE for enums, PascalCase for typed constants
- **Interfaces:** typically end with `-er` (e.g., `Cacher`, `Logger`)

### File Organization
```go
// 1. Package declaration
package handler

// 2. Imports (stdlib, then external, then internal)
import (
    "context"
    "time"

    "github.com/gofiber/fiber/v2"

    "github.com/duyet/gaxy/pkg/logger"
)

// 3. Constants and types
const DefaultTimeout = 30 * time.Second

type Handler struct { ... }

// 4. Constructor
func New(...) *Handler { ... }

// 5. Methods (public first, then private)
func (h *Handler) PublicMethod() { ... }
func (h *Handler) privateMethod() { ... }
```

### Documentation Standards

**Every exported symbol needs a godoc comment:**
```go
// Cache provides an in-memory LRU cache with TTL support.
// It automatically cleans up expired entries and enforces size limits.
type Cache struct { ... }

// Get retrieves an entry from the cache.
// Returns (entry, true) if found and not expired, (nil, false) otherwise.
func (c *Cache) Get(key string) (*Entry, bool) { ... }
```

### Testing Philosophy

**Write tests for:**
- All public APIs
- Error conditions
- Edge cases
- Configuration validation
- Critical business logic

**Test organization:**
```go
func TestServiceName(t *testing.T) {
    // Setup
    cfg := setupTestConfig()
    svc := NewService(cfg)

    // Execute
    result, err := svc.DoSomething()

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

## Development Workflow

### Before Coding
1. Understand the requirement fully
2. Check existing patterns in the codebase
3. Design the interface first
4. Consider error cases
5. Think about observability (logs, metrics)

### While Coding
1. Write self-documenting code
2. Add godoc comments
3. Include error context
4. Add metrics/logging as needed
5. Follow existing patterns

### After Coding
1. Run `make fmt` to format code
2. Run `make lint` to check for issues
3. Run `make test` to ensure tests pass
4. Update documentation if needed
5. Check metrics and logs work

### Commit Messages

Follow conventional commits:
```
feat: Add cache warming on startup
fix: Handle nil pointer in proxy service
docs: Update configuration reference
perf: Optimize cache lookup algorithm
refactor: Extract retry logic to client package
test: Add integration tests for rate limiting
```

## Common Tasks

### Adding a New Configuration Option

1. Add to `pkg/config/config.go`:
```go
type Config struct {
    // ...
    NewOption string `env:"NEW_OPTION" default:"value"`
}
```

2. Add validation in `Validate()`:
```go
if cfg.NewOption == "" {
    return errors.ValidationError("NEW_OPTION cannot be empty")
}
```

3. Update `.env.example`
4. Update README.md configuration table
5. Use in relevant service/handler

### Adding a New Metric

1. Define in `pkg/metrics/metrics.go`:
```go
type Metrics struct {
    // ...
    newMetric uint64
}
```

2. Add recording method:
```go
func (m *Metrics) RecordNewMetric(value int) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.newMetric += uint64(value)
}
```

3. Add to `Export()` method:
```go
output += fmt.Sprintf("# HELP gaxy_new_metric Description\n")
output += fmt.Sprintf("# TYPE gaxy_new_metric counter\n")
output += fmt.Sprintf("gaxy_new_metric %d\n\n", m.newMetric)
```

### Adding a New Middleware

1. Create in `pkg/middleware/middleware.go`:
```go
// NewMiddleware provides description of what it does
func NewMiddleware(deps Dependencies) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Pre-processing

        err := c.Next()

        // Post-processing

        return err
    }
}
```

2. Add to middleware stack in `server.go`:
```go
app.Use(middleware.NewMiddleware(deps))
```

3. Add tests
4. Document in README if user-facing

## Dependencies Philosophy

**When adding dependencies:**
- Prefer standard library when possible
- Choose well-maintained, popular libraries
- Check license compatibility (MIT preferred)
- Consider security implications
- Keep dependency tree shallow
- Pin versions in go.mod

**Current core dependencies:**
- `fiber/v2` - Fast HTTP framework
- `fasthttp` - High-performance HTTP client
- `envconfig` - Environment configuration
- `uuid` - Request ID generation
- `testify` - Testing utilities

## Deployment Considerations

### Docker
- Use multi-stage builds
- Minimize image size (distroless)
- Run as non-root user
- Include health checks
- Set resource limits

### Kubernetes
- Define resource requests/limits
- Configure liveness/readiness probes
- Use horizontal pod autoscaling
- Set up Prometheus scraping
- Configure log aggregation

### Monitoring
- Set up Grafana dashboards
- Configure alerting rules
- Monitor cache hit rates
- Track error rates
- Watch upstream latency

## Performance Tuning

### When performance issues arise:

1. **Profile first, optimize second**
```bash
go test -bench=. -benchmem -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

2. **Check these in order:**
   - Cache hit rate (should be >90% for static assets)
   - Connection pool settings
   - Upstream timeouts
   - Memory allocation patterns
   - Goroutine leaks

3. **Common optimizations:**
   - Increase cache size
   - Adjust connection pool limits
   - Use sync.Pool for allocations
   - Reduce logging verbosity
   - Enable compression

## Security Checklist

Before deploying to production:
- [ ] Rate limiting enabled
- [ ] Security headers configured
- [ ] CORS origins restricted (not *)
- [ ] HTTPS enforced
- [ ] Secrets not in environment variables
- [ ] Input validation on all endpoints
- [ ] Error messages don't leak internals
- [ ] Dependencies updated
- [ ] Container runs as non-root
- [ ] Resource limits configured

## Troubleshooting Guide

### High Memory Usage
1. Check cache size (`gaxy_cache_size_bytes`)
2. Look for goroutine leaks (`gaxy_info` with goroutine count)
3. Review connection pool settings
4. Check for memory leaks with pprof

### High Latency
1. Check upstream latency (`gaxy_upstream_duration_seconds`)
2. Review cache hit rate (`gaxy_cache_hits_total / gaxy_cache_misses_total`)
3. Check connection pool saturation
4. Look for rate limiting (`gaxy_rate_limit_dropped_total`)

### High Error Rate
1. Check logs for error patterns
2. Review upstream errors (`gaxy_upstream_errors_total`)
3. Verify configuration validity
4. Check rate limiting settings
5. Review security header impacts

## Future Enhancements

Ideas for future improvements:
- [ ] Circuit breaker for upstream failures
- [ ] Distributed caching (Redis)
- [ ] Request/response compression
- [ ] WebSocket support
- [ ] Multi-region routing
- [ ] A/B testing support
- [ ] Analytics dashboard
- [ ] Request replay for debugging
- [ ] Automated performance testing
- [ ] OpenTelemetry integration

## Questions to Ask

When making changes, always consider:
1. **Does this maintain backward compatibility?**
2. **How will this be monitored in production?**
3. **What happens when this fails?**
4. **Is this testable?**
5. **Will this scale?**
6. **Is this secure?**
7. **Is this documented?**

## Remember

> "Quality is not an act, it is a habit." - Aristotle

Every line of code is an opportunity to make the system better, more reliable, more observable, and more maintainable. Don't settle for "it works"—strive for "it's excellent."

---

**Last Updated:** 2025-11-16
**Maintainer:** Duyet (github.com/duyet)
**Contributors:** Claude (AI Assistant)
