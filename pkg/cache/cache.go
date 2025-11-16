package cache

import (
	"sync"
	"time"
)

// Entry represents a cached item
type Entry struct {
	Data      []byte
	ExpiresAt time.Time
	ContentType string
	StatusCode int
}

// Cache is a simple in-memory cache with TTL support
type Cache struct {
	mu       sync.RWMutex
	data     map[string]*Entry
	maxSize  int64
	currSize int64
	ttl      time.Duration
	stats    Stats
}

// Stats tracks cache statistics
type Stats struct {
	Hits       uint64
	Misses     uint64
	Evictions  uint64
	Sets       uint64
	Size       int64
	EntryCount int
}

// New creates a new cache with the specified TTL and max size
func New(ttl time.Duration, maxSize int64) *Cache {
	c := &Cache{
		data:    make(map[string]*Entry),
		maxSize: maxSize,
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go c.cleanupExpired()

	return c
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (*Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.stats.Misses++
		return nil, false
	}

	c.stats.Hits++
	return entry, true
}

// Set stores an item in the cache
func (c *Cache) Set(key string, data []byte, contentType string, statusCode int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	dataSize := int64(len(data))

	// Check if we need to evict items
	for c.currSize+dataSize > c.maxSize && len(c.data) > 0 {
		c.evictOldest()
	}

	entry := &Entry{
		Data:        data,
		ExpiresAt:   time.Now().Add(c.ttl),
		ContentType: contentType,
		StatusCode:  statusCode,
	}

	// Remove old entry if exists
	if oldEntry, exists := c.data[key]; exists {
		c.currSize -= int64(len(oldEntry.Data))
	}

	c.data[key] = entry
	c.currSize += dataSize
	c.stats.Sets++
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.data[key]; exists {
		c.currSize -= int64(len(entry.Data))
		delete(c.data, key)
	}
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*Entry)
	c.currSize = 0
}

// GetStats returns cache statistics
func (c *Cache) GetStats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = c.currSize
	stats.EntryCount = len(c.data)
	return stats
}

// evictOldest removes the oldest entry from the cache
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	first := true
	for key, entry := range c.data {
		if first || entry.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpiresAt
			first = false
		}
	}

	if oldestKey != "" {
		c.currSize -= int64(len(c.data[oldestKey].Data))
		delete(c.data, oldestKey)
		c.stats.Evictions++
	}
}

// cleanupExpired periodically removes expired entries
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.ExpiresAt) {
				c.currSize -= int64(len(entry.Data))
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}
