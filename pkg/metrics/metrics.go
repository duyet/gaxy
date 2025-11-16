package metrics

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// Metrics collects application metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	requestsTotal        map[string]uint64 // by status code
	requestDuration      []float64
	requestsInFlight     int64

	// Cache metrics
	cacheHits            uint64
	cacheMisses          uint64
	cacheEvictions       uint64
	cacheSizeBytes       int64
	cacheEntries         int64

	// Upstream metrics
	upstreamRequestsTotal   map[string]uint64 // by status code
	upstreamErrors          uint64
	upstreamDuration        []float64

	// Rate limit metrics
	rateLimitDropped     uint64

	startTime            time.Time
}

// New creates a new Metrics instance
func New() *Metrics {
	return &Metrics{
		requestsTotal:         make(map[string]uint64),
		upstreamRequestsTotal: make(map[string]uint64),
		requestDuration:       make([]float64, 0),
		upstreamDuration:      make([]float64, 0),
		startTime:             time.Now(),
	}
}

// RecordRequest records a request with its status code and duration
func (m *Metrics) RecordRequest(statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := strconv.Itoa(statusCode)
	m.requestsTotal[key]++
	m.requestDuration = append(m.requestDuration, duration.Seconds())

	// Keep only last 1000 durations for percentile calculation
	if len(m.requestDuration) > 1000 {
		m.requestDuration = m.requestDuration[len(m.requestDuration)-1000:]
	}
}

// IncRequestsInFlight increments the in-flight requests counter
func (m *Metrics) IncRequestsInFlight() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestsInFlight++
}

// DecRequestsInFlight decrements the in-flight requests counter
func (m *Metrics) DecRequestsInFlight() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestsInFlight--
}

// RecordUpstreamRequest records an upstream request
func (m *Metrics) RecordUpstreamRequest(statusCode int, duration time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := strconv.Itoa(statusCode)
	m.upstreamRequestsTotal[key]++
	m.upstreamDuration = append(m.upstreamDuration, duration.Seconds())

	if isError {
		m.upstreamErrors++
	}

	// Keep only last 1000 durations
	if len(m.upstreamDuration) > 1000 {
		m.upstreamDuration = m.upstreamDuration[len(m.upstreamDuration)-1000:]
	}
}

// UpdateCacheStats updates cache statistics
func (m *Metrics) UpdateCacheStats(hits, misses, evictions uint64, sizeBytes, entries int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheHits = hits
	m.cacheMisses = misses
	m.cacheEvictions = evictions
	m.cacheSizeBytes = sizeBytes
	m.cacheEntries = entries
}

// RecordRateLimitDrop records a rate-limited request
func (m *Metrics) RecordRateLimitDrop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimitDropped++
}

// Export exports metrics in Prometheus text format
func (m *Metrics) Export() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var output string

	// Process info
	output += fmt.Sprintf("# HELP gaxy_info Process information\n")
	output += fmt.Sprintf("# TYPE gaxy_info gauge\n")
	output += fmt.Sprintf("gaxy_info{version=\"1.0.0\"} 1\n\n")

	// Uptime
	uptime := time.Since(m.startTime).Seconds()
	output += fmt.Sprintf("# HELP gaxy_uptime_seconds Process uptime in seconds\n")
	output += fmt.Sprintf("# TYPE gaxy_uptime_seconds counter\n")
	output += fmt.Sprintf("gaxy_uptime_seconds %f\n\n", uptime)

	// Request metrics
	output += fmt.Sprintf("# HELP gaxy_requests_total Total number of HTTP requests\n")
	output += fmt.Sprintf("# TYPE gaxy_requests_total counter\n")
	for status, count := range m.requestsTotal {
		output += fmt.Sprintf("gaxy_requests_total{status=\"%s\"} %d\n", status, count)
	}
	output += "\n"

	// Requests in flight
	output += fmt.Sprintf("# HELP gaxy_requests_in_flight Number of requests currently being processed\n")
	output += fmt.Sprintf("# TYPE gaxy_requests_in_flight gauge\n")
	output += fmt.Sprintf("gaxy_requests_in_flight %d\n\n", m.requestsInFlight)

	// Request duration
	if len(m.requestDuration) > 0 {
		avg := average(m.requestDuration)
		p50 := percentile(m.requestDuration, 0.50)
		p95 := percentile(m.requestDuration, 0.95)
		p99 := percentile(m.requestDuration, 0.99)

		output += fmt.Sprintf("# HELP gaxy_request_duration_seconds HTTP request duration\n")
		output += fmt.Sprintf("# TYPE gaxy_request_duration_seconds summary\n")
		output += fmt.Sprintf("gaxy_request_duration_seconds{quantile=\"0.5\"} %f\n", p50)
		output += fmt.Sprintf("gaxy_request_duration_seconds{quantile=\"0.95\"} %f\n", p95)
		output += fmt.Sprintf("gaxy_request_duration_seconds{quantile=\"0.99\"} %f\n", p99)
		output += fmt.Sprintf("gaxy_request_duration_seconds_sum %f\n", avg*float64(len(m.requestDuration)))
		output += fmt.Sprintf("gaxy_request_duration_seconds_count %d\n\n", len(m.requestDuration))
	}

	// Cache metrics
	output += fmt.Sprintf("# HELP gaxy_cache_hits_total Total number of cache hits\n")
	output += fmt.Sprintf("# TYPE gaxy_cache_hits_total counter\n")
	output += fmt.Sprintf("gaxy_cache_hits_total %d\n\n", m.cacheHits)

	output += fmt.Sprintf("# HELP gaxy_cache_misses_total Total number of cache misses\n")
	output += fmt.Sprintf("# TYPE gaxy_cache_misses_total counter\n")
	output += fmt.Sprintf("gaxy_cache_misses_total %d\n\n", m.cacheMisses)

	output += fmt.Sprintf("# HELP gaxy_cache_evictions_total Total number of cache evictions\n")
	output += fmt.Sprintf("# TYPE gaxy_cache_evictions_total counter\n")
	output += fmt.Sprintf("gaxy_cache_evictions_total %d\n\n", m.cacheEvictions)

	output += fmt.Sprintf("# HELP gaxy_cache_size_bytes Current cache size in bytes\n")
	output += fmt.Sprintf("# TYPE gaxy_cache_size_bytes gauge\n")
	output += fmt.Sprintf("gaxy_cache_size_bytes %d\n\n", m.cacheSizeBytes)

	output += fmt.Sprintf("# HELP gaxy_cache_entries Current number of cache entries\n")
	output += fmt.Sprintf("# TYPE gaxy_cache_entries gauge\n")
	output += fmt.Sprintf("gaxy_cache_entries %d\n\n", m.cacheEntries)

	// Upstream metrics
	output += fmt.Sprintf("# HELP gaxy_upstream_requests_total Total number of upstream requests\n")
	output += fmt.Sprintf("# TYPE gaxy_upstream_requests_total counter\n")
	for status, count := range m.upstreamRequestsTotal {
		output += fmt.Sprintf("gaxy_upstream_requests_total{status=\"%s\"} %d\n", status, count)
	}
	output += "\n"

	output += fmt.Sprintf("# HELP gaxy_upstream_errors_total Total number of upstream errors\n")
	output += fmt.Sprintf("# TYPE gaxy_upstream_errors_total counter\n")
	output += fmt.Sprintf("gaxy_upstream_errors_total %d\n\n", m.upstreamErrors)

	// Upstream duration
	if len(m.upstreamDuration) > 0 {
		avg := average(m.upstreamDuration)
		p50 := percentile(m.upstreamDuration, 0.50)
		p95 := percentile(m.upstreamDuration, 0.95)
		p99 := percentile(m.upstreamDuration, 0.99)

		output += fmt.Sprintf("# HELP gaxy_upstream_duration_seconds Upstream request duration\n")
		output += fmt.Sprintf("# TYPE gaxy_upstream_duration_seconds summary\n")
		output += fmt.Sprintf("gaxy_upstream_duration_seconds{quantile=\"0.5\"} %f\n", p50)
		output += fmt.Sprintf("gaxy_upstream_duration_seconds{quantile=\"0.95\"} %f\n", p95)
		output += fmt.Sprintf("gaxy_upstream_duration_seconds{quantile=\"0.99\"} %f\n", p99)
		output += fmt.Sprintf("gaxy_upstream_duration_seconds_sum %f\n", avg*float64(len(m.upstreamDuration)))
		output += fmt.Sprintf("gaxy_upstream_duration_seconds_count %d\n\n", len(m.upstreamDuration))
	}

	// Rate limit metrics
	output += fmt.Sprintf("# HELP gaxy_rate_limit_dropped_total Total number of rate-limited requests\n")
	output += fmt.Sprintf("# TYPE gaxy_rate_limit_dropped_total counter\n")
	output += fmt.Sprintf("gaxy_rate_limit_dropped_total %d\n\n", m.rateLimitDropped)

	return output
}

// Helper functions for percentile calculation
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Simple percentile calculation (not sorting for performance)
	// In production, you might want to use a more efficient algorithm
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Simple bubble sort (sufficient for our use case with max 1000 elements)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * p)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}
