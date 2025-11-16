package ratelimit

import (
	"sync"
	"time"
)

// Limiter implements a token bucket rate limiter per IP
type Limiter struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
	rps     int
	burst   int
}

type bucket struct {
	tokens     float64
	lastUpdate time.Time
}

// New creates a new rate limiter
func New(rps, burst int) *Limiter {
	l := &Limiter{
		buckets: make(map[string]*bucket),
		rps:     rps,
		burst:   burst,
	}

	// Start cleanup goroutine
	go l.cleanup()

	return l
}

// Allow checks if a request from the given IP is allowed
func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	b, exists := l.buckets[ip]
	if !exists {
		b = &bucket{
			tokens:     float64(l.burst),
			lastUpdate: now,
		}
		l.buckets[ip] = b
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(b.lastUpdate).Seconds()
	b.tokens += elapsed * float64(l.rps)

	// Cap at burst limit
	if b.tokens > float64(l.burst) {
		b.tokens = float64(l.burst)
	}

	b.lastUpdate = now

	// Check if we have tokens available
	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}

	return false
}

// cleanup removes old buckets to prevent memory leak
func (l *Limiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, b := range l.buckets {
			// Remove buckets that haven't been accessed in 5 minutes
			if now.Sub(b.lastUpdate) > 5*time.Minute {
				delete(l.buckets, ip)
			}
		}
		l.mu.Unlock()
	}
}
