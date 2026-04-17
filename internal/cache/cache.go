package cache

import (
	"sync"
	"time"
)

// Entry holds a cached value with expiration metadata.
type Entry[T any] struct {
	Value     T
	FetchedAt time.Time
	ExpiresAt time.Time
}

// Cache is a generic in-memory TTL cache with stale-while-revalidate support.
type Cache[T any] struct {
	mu      sync.RWMutex
	entries map[string]*Entry[T]
	ttl     time.Duration
}

// New creates a cache with the given TTL.
func New[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		entries: make(map[string]*Entry[T]),
		ttl:     ttl,
	}
}

// Get returns the cached value and whether it's still fresh.
// Returns (value, true, true) if fresh, (value, true, false) if stale, (zero, false, false) if missing.
func (c *Cache[T]) Get(key string) (T, bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		var zero T
		return zero, false, false
	}
	fresh := time.Now().Before(entry.ExpiresAt)
	return entry.Value, true, fresh
}

// Set stores a value in the cache.
func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.entries[key] = &Entry[T]{
		Value:     value,
		FetchedAt: now,
		ExpiresAt: now.Add(c.ttl),
	}
}
