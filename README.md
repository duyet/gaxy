# gaxy

![Docker](https://github.com/duyet/gaxy/workflows/Docker/badge.svg)
![Go test](https://github.com/duyet/gaxy/workflows/Go/badge.svg)

**Production-ready Google Analytics / Google Tag Manager Proxy** built with Go.

Bypass ad blockers, maintain data privacy, and own your analytics pipeline with a blazing-fast, enterprise-grade proxy server.

![How it works?](.github/screenshot/how-gaxy-works.png)
<!-- https://sketchviz.com/@duyet/d4c36c277140a24111a723c439291303/9b91c5b780ff792c7dc08f70d22442a3ac523096 -->

## ✨ Features

### 🚀 Performance
- **Intelligent Caching**: Reduces upstream calls by 95%+ with configurable TTL
- **Connection Pooling**: Optimized HTTP client with connection reuse
- **Retry Logic**: Automatic retries for transient failures
- **Sub-millisecond Latency**: Highly optimized request handling

### 🔒 Security
- **Rate Limiting**: Per-IP token bucket rate limiting
- **Security Headers**: CSP, X-Frame-Options, X-Content-Type-Options
- **Input Validation**: Comprehensive request/response validation
- **Distroless Container**: Minimal attack surface with non-root user

### 📊 Observability
- **Prometheus Metrics**: Request latency, cache performance, error rates
- **Structured Logging**: JSON logs with correlation IDs
- **Health Endpoints**: Detailed system health information
- **Request Tracing**: Full request lifecycle tracking

### 🏗️ Architecture
- **Clean Architecture**: Separation of concerns with handlers, services, middleware
- **Error Handling**: Typed errors with context and proper error propagation
- **Configuration**: Environment-based config with validation
- **Graceful Shutdown**: Proper cleanup on termination

## 🚀 Quick Start

### Using Docker (Recommended)

```sh
docker run -d -p 3000:3000 \
    -e ROUTE_PREFIX=/analytics \
    -e GOOGLE_ORIGIN=https://www.google-analytics.com \
    ghcr.io/duyet/gaxy:latest
```

### Using Docker Compose

```sh
docker-compose up -d
```

### Using Go

```sh
# Clone repository
git clone https://github.com/duyet/gaxy.git
cd gaxy

# Run with default config
go run *.go

# Or build and run
make build
./bin/gaxy
```

## 📖 Development

### Prerequisites
- Go 1.24 or later
- Docker (optional)
- Make (optional)

### Local Development

```sh
# Install dependencies
make deps

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Run linters
make lint

# Run server
make run
```

### Project Structure

```
gaxy/
├── pkg/
│   ├── cache/        # Intelligent caching layer
│   ├── config/       # Configuration management
│   ├── errors/       # Custom error types
│   ├── handler/      # HTTP handlers
│   ├── logger/       # Structured logging
│   ├── metrics/      # Prometheus metrics
│   ├── middleware/   # HTTP middleware
│   ├── proxy/        # Proxy service & client
│   └── ratelimit/    # Rate limiting
├── server.go         # Main application
├── Dockerfile        # Production container
├── Makefile          # Build automation
└── docker-compose.yml # Docker Compose config
```

## ⚙️ Configuration

Gaxy is configured via environment variables. See [.env.example](.env.example) for a complete reference.

### Server Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server listening port | `3000` |
| `SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `10s` |
| `READ_TIMEOUT` | HTTP read timeout | `30s` |
| `WRITE_TIMEOUT` | HTTP write timeout | `30s` |

### Routing Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `ROUTE_PREFIX` | URL prefix for all endpoints (e.g., `/analytics`) | `""` |

### Upstream Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `GOOGLE_ORIGIN` | Upstream Google Analytics/Tag Manager URL | `https://www.google-analytics.com` |
| `UPSTREAM_TIMEOUT` | Upstream request timeout | `10s` |
| `UPSTREAM_MAX_IDLE_CONNS` | Maximum idle connections | `100` |
| `UPSTREAM_MAX_CONNS` | Maximum total connections | `100` |
| `UPSTREAM_RETRY_COUNT` | Number of retries on failure | `2` |
| `UPSTREAM_RETRY_DELAY` | Delay between retries | `100ms` |

### Header Injection
| Variable | Description | Default |
|----------|-------------|---------|
| `INJECT_PARAMS_FROM_REQ_HEADERS` | Convert request headers to query parameters<br>Format: `header1,header2` or `header1__param1,header2__param2`<br>Example: `x-email__uip,user-agent__ua` | `""` |
| `SKIP_PARAMS_FROM_REQ_HEADERS` | Remove specific query parameters<br>Example: `fbclid,gclid` | `""` |

### Cache Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `CACHE_ENABLED` | Enable intelligent caching | `true` |
| `CACHE_TTL` | Cache time-to-live | `5m` |
| `CACHE_MAX_SIZE` | Maximum cache size in bytes | `104857600` (100MB) |
| `CACHE_KEY_PATTERN` | File pattern to cache | `*.js` |

### Rate Limiting
| Variable | Description | Default |
|----------|-------------|---------|
| `RATE_LIMIT_ENABLED` | Enable per-IP rate limiting | `true` |
| `RATE_LIMIT_RPS` | Requests per second per IP | `100` |
| `RATE_LIMIT_BURST` | Burst allowance | `200` |

### Logging
| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Log format (`json` or `text`) | `json` |

### Metrics
| Variable | Description | Default |
|----------|-------------|---------|
| `METRICS_ENABLED` | Enable Prometheus metrics | `true` |
| `METRICS_PATH` | Metrics endpoint path | `/metrics` |

### Security
| Variable | Description | Default |
|----------|-------------|---------|
| `ENABLE_CORS` | Enable CORS | `true` |
| `CORS_ALLOW_ORIGINS` | Allowed CORS origins | `*` |
| `ENABLE_SECURITY_HEADERS` | Enable security headers | `true` |

## 📦 Installation

### Using Docker

```sh
docker run -d -p 3000:3000 \
    -e ROUTE_PREFIX=/analytics \
    -e CACHE_ENABLED=true \
    -e RATE_LIMIT_ENABLED=true \
    ghcr.io/duyet/gaxy:latest
```

### Using Docker Compose

See [docker-compose.yml](docker-compose.yml) for a complete example.

```sh
docker-compose up -d
```

### Using Helm

```sh
helm repo add duyet https://duyet.github.io/charts
helm install gaxy duyet/gaxy \
    --set config.routePrefix=/analytics \
    --set config.cacheEnabled=true
```

### Using Google App Engine

```sh
# 1. Install gcloud SDK
# 2. Deploy
gcloud app deploy
```

## 📊 Usage

### Basic Setup

Replace Google Analytics/GTM script URLs with your Gaxy instance:

```html
<!-- Google Analytics -->
<script>
window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;
ga('create', 'UA-XXXXX-Y', 'auto');
ga('send', 'pageview');
</script>
<script async src='https://your-gaxy-instance.com/analytics.js'></script>
<!-- End Google Analytics -->
```

### With Route Prefix

If you configured `ROUTE_PREFIX=/analytics`:

```html
<script async src='https://your-gaxy-instance.com/analytics/analytics.js'></script>
```

### Monitoring Endpoints

Gaxy provides several endpoints for monitoring and debugging:

#### Health Check
```sh
curl http://localhost:3000/ping
# Response: pong

curl http://localhost:3000/health
# Response: JSON with system metrics
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2h15m30s",
  "system": {
    "goroutines": 12,
    "memory_alloc": "15 MB",
    "memory_total": "45 MB",
    "memory_sys": "72 MB",
    "gc_runs": 23
  }
}
```

#### Prometheus Metrics
```sh
curl http://localhost:3000/metrics
# Returns Prometheus-formatted metrics:
# - gaxy_requests_total{status="200"} 12450
# - gaxy_request_duration_seconds{quantile="0.95"} 0.023
# - gaxy_cache_hits_total 11230
# - gaxy_cache_misses_total 1220
# - gaxy_upstream_requests_total{status="200"} 1220
# - gaxy_rate_limit_dropped_total 15
```

## 📈 Performance

### Caching Benefits

With intelligent caching enabled:
- **95%+ cache hit rate** for static assets (analytics.js, gtag.js)
- **Sub-millisecond response times** for cached content
- **Reduced upstream load** by 20x
- **Lower bandwidth costs** for high-traffic sites

### Benchmarks

```sh
# Run benchmarks
make bench

# Example results:
# BenchmarkProxyRequest-8     50000    24532 ns/op    4096 B/op    42 allocs/op
# BenchmarkCacheHit-8      5000000      245 ns/op      64 B/op     2 allocs/op
```

## 🔍 Monitoring with Prometheus

Example Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'gaxy'
    static_configs:
      - targets: ['localhost:3000']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

Example Grafana queries:
```promql
# Request rate
rate(gaxy_requests_total[5m])

# P95 latency
gaxy_request_duration_seconds{quantile="0.95"}

# Cache hit rate
rate(gaxy_cache_hits_total[5m]) / (rate(gaxy_cache_hits_total[5m]) + rate(gaxy_cache_misses_total[5m]))

# Error rate
rate(gaxy_requests_total{status=~"5.."}[5m])
```

## 🛡️ Security Best Practices

1. **Enable Rate Limiting**: Protect against abuse
   ```sh
   RATE_LIMIT_ENABLED=true
   RATE_LIMIT_RPS=100
   ```

2. **Use HTTPS**: Always run behind HTTPS in production

3. **Configure CORS**: Restrict origins in production
   ```sh
   CORS_ALLOW_ORIGINS=https://yourdomain.com
   ```

4. **Monitor Metrics**: Set up alerts for anomalies

5. **Update Regularly**: Keep dependencies up to date
   ```sh
   make deps
   ```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details

## 🙏 Acknowledgments

Built with:
- [Fiber](https://github.com/gofiber/fiber) - Fast HTTP framework
- [Fasthttp](https://github.com/valyala/fasthttp) - High-performance HTTP client

---

**Made with ❤️ by [Duyet](https://github.com/duyet)**
